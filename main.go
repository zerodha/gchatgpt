package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/lestrrat-go/jwx/jwk"
)

const (
	OPENAI_API_URL = "https://api.openai.com/v1/chat/completions"
	GOOGLE_JWT_URL = "https://www.googleapis.com/service_accounts/v1/jwk/chat@system.gserviceaccount.com"
)

var (
	hc    *http.Client
	lo    = log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lshortfile)
	jwSet jwk.Set

	model            = "gpt-3.5-turbo"
	openaiAPIKey     = ""
	googleProjectNum = ""
	prePrompt        = ""
)

type InReq struct {
	Text string `json:"text"`
}

type APIReq struct {
	Model   string `json:"model"`
	Message []struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	} `json:"messages"`
}

type APIResp struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("send a POST request"))
		return
	}

	// Validate JWT header.
	h := r.Header.Get("Authorization")
	if !strings.HasPrefix(h, "Bearer") {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("missing authorization header"))
		return
	}
	if err := verifyToken(strings.TrimPrefix(h, "Bearer "), googleProjectNum); err != nil {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(err.Error()))
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		lo.Printf("error reading request body: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var data InReq
	if err := json.Unmarshal(body, &data); err != nil {
		lo.Printf("error unmarshalling request: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if data.Text == "" {
		lo.Printf("missing text in request: %s", body)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	resp, err := queryAPI(data.Text, prePrompt, model, openaiAPIKey)
	if err != nil {
		lo.Printf("error querying GPT API: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	b, err := json.Marshal(InReq{Text: resp})
	if err != nil {
		lo.Printf("error marshalling GPT response: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(b)
}

func queryAPI(prompt, prePrompt, model, apiKey string) (string, error) {
	payload := APIReq{
		Model: model,
		Message: []struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		}{
			{Role: "user", Content: prePrompt + "\n"},
			{Role: "user", Content: prompt},
		},
	}

	b, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	// Prepare the request.
	req, err := http.NewRequest(http.MethodPost, OPENAI_API_URL, bytes.NewReader(b))
	if err != nil {
		return "", err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", apiKey))

	// Make the request.
	resp, err := hc.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Read the response.
	var data APIResp
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", err
	}

	if len(data.Choices) == 0 {
		return "", fmt.Errorf("no answers from the API")
	}

	return data.Choices[0].Message.Content, nil
}

func verifyToken(tokenString, projectNum string) error {
	token, err := jwt.Parse(tokenString, getKey)
	if err != nil {
		return err
	}

	claims := token.Claims.(jwt.MapClaims)
	for key, value := range claims {
		// "Audience" / project ID.
		if key == "aud" && value == projectNum {
			return nil
		}
	}

	return errors.New("failed to authenticate message")
}

func getKey(token *jwt.Token) (interface{}, error) {
	keyID, ok := token.Header["kid"].(string)
	if !ok {
		return nil, errors.New("expecting JWT header to have string kid")
	}

	if key, ok := jwSet.LookupKeyID(keyID); ok {
		var pubkey interface{}
		err := key.Raw(&pubkey)
		return pubkey, err
	}

	return nil, fmt.Errorf("unable to find key %q", keyID)
}

func main() {
	var (
		fAddr       = flag.String("addr", ":8000", "address to listen on")
		fReqTimeout = flag.Duration("timeout", time.Second*10, "API request timeout")
		fConnPool   = flag.Int("conn-pool", 10, "HTTP connection pool")
		fPrePrompt  = flag.String("pre-prompt", "You are a non-verbose and to-the-point agent",
			"default prompt to append before user chat queries")
		fModel = flag.String("openai-model", "gpt-3.5-turbo", "OpenAI model name to use")

		// Google chat app's numeric project/client ID
		fGoogProjectNum = os.Getenv("GOOGLE_PROJECT_NUMBER")
		fOpenaiApiKey   = os.Getenv("OPENAI_API_KEY")
	)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "set env vars: OPENAI_API_KEY, GOOGLE_PROJECT_NUMBER\n\n")
		flag.VisitAll(func(f *flag.Flag) {
			fmt.Fprintf(os.Stderr, "  -%s  %v\n", f.Name, f.Usage) // f.Name, f.Value
		})
	}
	flag.Parse()

	if fOpenaiApiKey == "" || fGoogProjectNum == "" {
		lo.Fatalf("pass env variables OPENAI_API_KEY, GOOGLE_API_KEY")
	}

	// Set globals config.
	openaiAPIKey = fOpenaiApiKey
	googleProjectNum = fGoogProjectNum
	prePrompt = *fPrePrompt
	model = *fModel

	// Fetch Google's JWK.
	if j, err := jwk.Fetch(context.TODO(), GOOGLE_JWT_URL); err != nil {
		log.Fatalf("error fetching Google JWK: %s: %v", GOOGLE_JWT_URL, err)
	} else {
		jwSet = j
	}

	// Initialize the global HTTP client.
	hc = &http.Client{
		Timeout: *fReqTimeout,
		Transport: &http.Transport{
			MaxIdleConnsPerHost:   *fConnPool,
			ResponseHeaderTimeout: *fReqTimeout,
		},
	}

	// Start the webserver.
	http.HandleFunc("/", handleRequest)
	lo.Printf("listening on %s", *fAddr)

	err := http.ListenAndServe(*fAddr, nil)
	lo.Fatalf("error starting HTTP server: %v", err)
}
