package server

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/zerodhatech/gchatgpt/pkg/gchat"
)

// HandleGChat handles the Google Chat webhook.
func (app *App) HandleGChat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	bearerToken := r.Header.Get("Authorization")
	token := strings.Split(bearerToken, " ")
	err := app.cfg.JWKVerifier.VerifyJWT(app.cfg.BotAppID, token[1])
	if len(token) != 2 || err != nil {
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}

	message := &gchat.Event{}
	err = json.NewDecoder(r.Body).Decode(&message)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Println(err)
		return
	}

	// If the bot is added to a space, respond with a welcome message.
	if message.Type == "ADDED_TO_SPACE" {
		response := gchat.Response{Text: "Thanks for adding me!"}
		json.NewEncoder(w).Encode(response)
		return
	}

	// If the message type is not a message, ignore it.
	if message.Type != "MESSAGE" {
		response := gchat.Response{Text: "Sorry, I didn't understand your message."}
		json.NewEncoder(w).Encode(response)
		return
	}

	userName := message.Message.Sender.DisplayName

	// Cleanup the prompt message. Argument text has a leading space usually.
	prompt := fmt.Sprintf("%s: %s", userName, strings.TrimSpace(message.Message.ArgumentText))

	interactionKey := getInteractionKey(message)

	// Send the prompt to OpenAI and get a response.
	response, err := app.cfg.OpenAI.RespondWithPrompt(
		interactionKey,
		app.cfg.PrePrompt,
		prompt,
	)

	if err != nil {
		log.Println(err)
		response := gchat.Response{Text: "Sorry, I didn't understand your message."}
		json.NewEncoder(w).Encode(response)
		return
	}

	out := gchat.Response{Text: response}
	if err := json.NewEncoder(w).Encode(out); err != nil {
		log.Println(err)
		response := gchat.Response{Text: "Sorry, I didn't understand your message."}
		json.NewEncoder(w).Encode(response)
		return
	}
}

// getInteractionKey returns the interaction key for the message. TODO: This is
// a temporary implementation. We need to come up with a better way to
// identify the interaction.
func getInteractionKey(message *gchat.Event) string {
	switch message.Message.Space.Type {
	case gchat.SpaceTypeDM:
		return message.Message.Space.Name
	case gchat.SpaceTypeRoom:
		return message.Message.Thread.Name
	default:
		// Return a generated random string for now.
		return randomString(10)
	}
}

func randomString(length int) string {
	// Define the character set from which to generate the random string
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	// Seed the random number generator
	rand.Seed(time.Now().UnixNano())

	// Create a byte slice of the specified length
	randomBytes := make([]byte, length)

	// Fill the byte slice with random characters from the character set
	for i := 0; i < length; i++ {
		randomBytes[i] = charset[rand.Intn(len(charset))]
	}

	// Convert the byte slice to a string and return it
	return string(randomBytes)
}
