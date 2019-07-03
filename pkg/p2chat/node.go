package p2chat

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/protocol"
	"github.com/libp2p/go-libp2p-pubsub"

	"github.com/multiformats/go-multiaddr"
)

type Node struct {
	host    host.Host
	pubsub  *pubsub.PubSub
	pbMutex sync.Mutex
}

func NewNode(ctx context.Context, addr multiaddr.Multiaddr, privKey crypto.PrivKey, protocolID string) (*Node, error) {
	host, err := libp2p.New(
		ctx,
		libp2p.ListenAddrs(addr),
		libp2p.Identity(privKey),
	)
	if err != nil {
		return nil, err
	}

	pb, err := pubsub.NewFloodsubWithProtocols(
		ctx,
		host,
		[]protocol.ID{protocol.ID(protocolID)},
		pubsub.WithMessageSigning(true),
		pubsub.WithStrictSignatureVerification(true),
	)
	if err != nil {
		return nil, err
	}

	node := &Node{
		host:   host,
		pubsub: pb,
	}

	return node, nil
}

func (node *Node) Subscribe(ctx context.Context, topic string, handler Handler) error {
	return newSubscription(ctx, node, topic, handler)
}

func (node *Node) Publish(topic string, data interface{}) error {
	sendData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	node.pbMutex.Lock()
	defer node.pbMutex.Unlock()

	if err := node.pubsub.Publish(topic, sendData); err != nil {
		return err
	}

	return nil
}

func (node *Node) Close() error {
	return node.host.Close()
}

func (node *Node) HostID() string {
	return node.host.ID().Pretty()
}
