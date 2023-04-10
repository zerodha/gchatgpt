package gchat

const (
	SpaceTypeDM   = "DM"
	SpaceTypeRoom = "ROOM"
)

// Event model corresponds to the event sent by Google Chat.
type Event struct {
	Type    string `json:"type"`
	Message struct {
		Text         string `json:"text"`
		ArgumentText string `json:"argumentText"`
		Sender       struct {
			DisplayName string `json:"displayName"`
		} `json:"sender"`
		Thread struct {
			Name string `json:"name"`
		} `json:"thread"`
		Space struct {
			Name string `json:"name"`
			Type string `json:"type"`
		} `json:"space"`
	} `json:"message"`
}

// Response model corresponds to the response sent by the bot.
type Response struct {
	Text string `json:"text"`
}
