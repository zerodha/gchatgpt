package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/zerodhatech/gchatgpt/pkg/gchat"
	"github.com/zerodhatech/gchatgpt/pkg/openai"
	"github.com/zerodhatech/gchatgpt/pkg/server"
)

var (
	apiOpenAIKey = os.Getenv("OPENAI_API_KEY")
	botAppID     = os.Getenv("GOOGLE_PROJECT_NUMBER")
	address      = os.Getenv("ADDRESS")
	gptModel     = os.Getenv("GPT_MODEL")
	prePrompt    = os.Getenv("PRE_PROMPT")
)

func main() {
	// TODO: setup koanf to read flags and other ways to configure

	// TODO: setup http client
	httpClient := &http.Client{}

	// Setup openai client
	client, err := openai.NewClient(openai.ClientConfig{
		HTTPClient: httpClient,
		APIKey:     apiOpenAIKey,
		RootURL:    openai.APIURLv1,
	})
	if err != nil {
		log.Fatalf("error creating openai client: %v", err)
	}

	// Setup openai store. TODO: make configurable
	ttl := int64(24 * time.Hour) // 1 day
	maxHistory := 10             // store only 10

	store := openai.NewStore(ttl, maxHistory)

	// Setup openai responder utility
	openAI := openai.NewResponder(openai.ResponderOptions{
		Client: client,
		Store:  store,
		Model:  gptModel,
	})

	// Setup jwk verifier
	jwkVerifier, err := gchat.NewJWKVerifier()
	if err != nil {
		log.Fatalf("error creating jwk verifier: %v", err)
	}

	cfg := server.Options{
		OpenAIKey:   apiOpenAIKey,
		BotAppID:    botAppID,
		Address:     address,
		OpenAI:      openAI,
		JWKVerifier: jwkVerifier,
		PrePrompt:   prePrompt,
	}

	app := server.New(cfg)

	log.Fatal(app.Run())
}
