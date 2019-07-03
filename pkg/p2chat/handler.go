package p2chat

import (
	"github.com/libp2p/go-libp2p-core/peer"
)

type PeerAddr peer.AddrInfo

type Handler interface {
	TextMessage(msg TextMessage)
	Peer(addr PeerAddr)
}

// TextMessage is more end-user model of regular text messages
type TextMessage struct {
	Body string
	From peer.ID
}
