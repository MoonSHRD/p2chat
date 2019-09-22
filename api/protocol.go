package api

import "github.com/libp2p/go-libp2p-core/peer"

/*
Flags:
		- 0x0: Generic message
		- 0x1: Request to get existing PubSub topics at the network
		- 0x2: Response to the request for topics (ack)
		- 0x3: Request to ask peers for their MatrixID
		- 0x4: Response to the request for MatrixID
*/
const (
	FlagGenericMessage   int = 0x0
	FlagTopicsRequest    int = 0x1
	FlagTopicsResponse   int = 0x2
	FlagIdentityRequest  int = 0x3
	FlagIdentityResponse int = 0x4
)

// BaseMessage is the basic message format of our protocol
type BaseMessage struct {
	Body string `json:"body"`
	To   string `json:"to"`
	Flag int    `json:"flag"`
}

// GetTopicsRespondMessage is the format of the message to answer of request for topics
// Flag: 0x2
type GetTopicsRespondMessage struct {
	BaseMessage
	Topics []string `json:"topics"`
}

// GetIdentityRespondMessage is the format of the message to answer of request for peer identity
// Flag: 0x4
type GetIdentityRespondMessage struct {
	BaseMessage
	PeerID   peer.ID `json:"peer_id"`
	MatrixID string  `json:"matrix_id"`
}
