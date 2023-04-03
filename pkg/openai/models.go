package openai

const (
	APIURLv1               = "https://api.openai.com/v1"
	chatCompletionEndpoint = "/chat/completions"

	RoleSystem    = "system"
	RoleUser      = "user"
	RoleAssistant = "assistant"
)

// ChatCompletionMessage is a message in a chat completion request.
type ChatCompletionMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatCompletionRequest is the request for the chat completion endpoint.
type ChatCompletionRequest struct {
	Model    string                  `json:"model"`
	Messages []ChatCompletionMessage `json:"messages"`
}

// ChatCompletionResponse is the response from the chat completion endpoint.
type ChatCompletionResponse struct {
	Choices []struct {
		Message ChatCompletionMessage `json:"message"`
	} `json:"choices"`
}
