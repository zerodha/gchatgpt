package openai

import (
	"sync"
	"time"
)

const (
	defaultModel = "gpt-3.5-turbo"
)

// Store is a thread-safe map with a TTL.
type Store struct {
	cache      sync.Map
	ttl        int64
	maxHistory int
}

// StoreValue is the value stored in the Store.
type StoreValue struct {
	Messages []ChatCompletionMessage
	TTL      int64
}

// NewStore creates a new Store.
func NewStore(ttl int64, maxHistory int) *Store {
	s := &Store{
		ttl:        ttl,
		maxHistory: maxHistory,
	}

	go func() {
		for {
			s.cleanup()
			time.Sleep(time.Hour * 1) // TODO: make configurable, hardcoded to 1 hour
		}
	}()

	return s
}

func (s *Store) cleanup() {
	s.cache.Range(func(key, val interface{}) bool {
		v, ok := val.(StoreValue)
		if !ok {
			return true
		}

		if v.TTL < time.Now().Unix() {
			s.cache.Delete(key)
		}

		return true
	})
}

// Get gets the value from the Store.
func (s *Store) Get(key string) ([]ChatCompletionMessage, bool) {
	val, ok := s.cache.Load(key)
	if !ok {
		return nil, false
	}

	v, ok := val.(StoreValue)
	if !ok {
		return nil, false
	}

	if v.TTL < time.Now().Unix() {
		s.cache.Delete(key)
		return nil, false
	}

	return v.Messages, true
}

// Set sets the value in the Store.
func (s *Store) Set(key string, val []ChatCompletionMessage) {
	s.cache.Store(key, StoreValue{
		Messages: val,
		TTL:      time.Now().Unix() + s.ttl,
	})
}

// ResponderOptions are the options for the Responder.
type ResponderOptions struct {
	Client *Client
	Store  *Store

	Model string
}

// Responder is a wrapper around the OpenAI API.
type Responder struct {
	client *Client
	store  *Store
	model  string
}

// NewResponder creates a new Responder.
func NewResponder(opt ResponderOptions) *Responder {
	if opt.Model == "" {
		opt.Model = defaultModel
	}

	return &Responder{
		client: opt.Client,
		store:  opt.Store,
		model:  opt.Model,
	}
}

// Respond responds to the prompts. InteractionKey is used to cache the
// responses.
func (r *Responder) Respond(interactionKey string, prompts []ChatCompletionMessage) (string, error) {
	// check if we have a cached response
	if msgs, ok := r.store.Get(interactionKey); ok {
		prompts = append(msgs, prompts...)
	}

	resp, err := r.client.ChatCompletion(ChatCompletionRequest{
		Model:    r.model,
		Messages: prompts,
	})
	if err != nil {
		return "", err
	}

	// cache the response
	msgsRecv := []ChatCompletionMessage{}
	for _, msg := range resp.Choices {
		msgsRecv = append(msgsRecv, msg.Message)
	}

	// if cache size is greater than 10, remove the oldest message
	if len(msgsRecv) > r.store.maxHistory {
		msgsRecv = msgsRecv[1:]
	}

	r.store.Set(interactionKey, msgsRecv)

	return resp.Choices[0].Message.Content, nil
}

// RespondWithPrompt responds to the prompts. InteractionKey is used to cache the
// responses. It is a utility function that adds the prompt to the prompts.
func (r *Responder) RespondWithPrompt(interactionKey, prePrompt, prompt string) (string, error) {
	prompts := []ChatCompletionMessage{}
	if prePrompt != "" {
		prompts = append(prompts, ChatCompletionMessage{
			Role:    RoleSystem,
			Content: prePrompt,
		})
	}

	prompts = append(prompts, ChatCompletionMessage{
		Role:    RoleUser,
		Content: prompt,
	})

	return r.Respond(interactionKey, prompts)
}
