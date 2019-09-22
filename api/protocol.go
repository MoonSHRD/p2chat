package api

/*
Flags:
		- 0x0: Generic message
		- 0x1: Request to get existing PubSub topics at the network
		- 0x2: Response to the request for topics (ack)
		- 0x3: Request to ask peers for their MatrixID
		- 0x4: Response to the request for MatrixID
		- 0x5: Greeting to users in the topic
		- 0x6: Farewell to users in the topic
		- 0x7: Same as 0x4, but for response to greeting in the topic
*/
const (
	FlagGenericMessage   int = 0x0
	FlagTopicsRequest    int = 0x1
	FlagTopicsResponse   int = 0x2
	FlagIdentityRequest  int = 0x3
	FlagIdentityResponse int = 0x4
	FlagGreeting         int = 0x5
	FlagFarewell         int = 0x6
	FlagGreetingRespond  int = 0x7

	ProtocolString string = "/moonshard/2.0.0"
)

// BaseMessage is the basic message format of our protocol
type BaseMessage struct {
	Body         string `json:"body"`
	To           string `json:"to"`
	Flag         int    `json:"flag"`
	FromMatrixID string `json:"fromMatrixID"`
}

// GetTopicsRespondMessage is the format of the message to answer of request for topics
// Flag: 0x2
type GetTopicsRespondMessage struct {
	BaseMessage
	Topics []string `json:"topics"`
}
