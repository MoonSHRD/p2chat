package api

/*
	The basic message format of our protocol
	Flags:
		- 0x0: Generic message
		- 0x1: Request to get existing PubSub topics on the network
		- 0x2: Answer to the request for topics (ack)
*/
type BaseMessage struct {
	Body string `json:"body"`
	Flag int    `json:"flag"`
}

/*

	The format of the message to answer of request for topics
	Flag: 0x2
*/
type GetTopicsAckMessage struct {
	BaseMessage
	Topics []string `json:"topics"`
}