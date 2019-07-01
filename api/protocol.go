package api

/*
Flags:
		- 0x0: Generic message
		- 0x1: Request to get existing PubSub topics at the network
		- 0x2: Response to the request for topics (ack)
*/
const (
	FLAG_GENERIC_MESSAGE int = 0x0
	FLAG_TOPICS_REQUEST  int = 0x1
	FLAG_TOPICS_RESPONSE int = 0x2
)

// BaseMessage is the basic message format of our protocol
type BaseMessage struct {
	Body string `json:"body"`
	Flag int    `json:"flag"`
}

// GetTopicsRespondMessage is the format of the message to answer of request for topics
// Flag: 0x2
type GetTopicsRespondMessage struct {
	BaseMessage
	Topics []string `json:"topics"`
}
