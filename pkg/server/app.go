package server

import (
	"log"
	"net/http"

	"github.com/zerodhatech/gchatgpt/pkg/gchat"
	"github.com/zerodhatech/gchatgpt/pkg/openai"
)

const (
	defaultPrePrompt = "You are a chat bot in a thread with multiple users. You will receive messages in the format <username>:<prompt>. Respond in a non-verbose and to-the-point manner."
)

// Options contains the configuration for the server.
type Options struct {
	OpenAIKey string
	BotAppID  string
	PrePrompt string

	Address string

	OpenAI      *openai.Responder
	JWKVerifier *gchat.JWKVerifier
}

// App is the server.
type App struct {
	cfg Options
}

// New creates a new server.
func New(cfg Options) *App {
	if cfg.Address == "" {
		cfg.Address = ":1234"
	}

	if cfg.PrePrompt == "" {
		cfg.PrePrompt = defaultPrePrompt
	}

	return &App{
		cfg: cfg,
	}
}

// Run starts the server.
func (app *App) Run() error {
	http.HandleFunc("/", app.HandleGChat)

	log.Println("Starting server on ", app.cfg.Address)

	return http.ListenAndServe(app.cfg.Address, nil)
}
