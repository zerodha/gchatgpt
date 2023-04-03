package gchat

// Event model corresponds to the event sent by Google Chat.
type Event struct {
	Type    string `json:"type"`
	Message struct {
		Text         string `json:"text"`
		ArgumentText string `json:"argumentText"`
		Thread       struct {
			Name string `json:"name"`
		}
	} `json:"message"`
}

// Response model corresponds to the response sent by the bot.
type Response struct {
	Text string `json:"text"`
}
