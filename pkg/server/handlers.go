package server

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

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

	// Cleanup the prompt message. Argument text has a leading space usually.
	prompt := strings.TrimSpace(message.Message.ArgumentText)

	interactionKey := getInteractionKey(message)

	// Send the prompt to OpenAI and get a response.
	response, err := app.cfg.OpenAI.RespondWithPrompt(
		interactionKey,
		"",
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
	return message.Message.Thread.Name
}
