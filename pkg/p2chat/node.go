package p2chat

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/MoonSHRD/p2chat/api"
	"github.com/MoonSHRD/p2chat/pkg/mdns"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/peerstore"
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
	incomingMessages := make(chan pubsub.Message)

	subscription, err := node.pubsub.Subscribe(topic)
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}
			msg, err := subscription.Next(ctx)
			if err != nil {
				fmt.Println("Error reading from buffer")
				return
			}

			if len(msg.Data) == 0 {
				continue
			}
			addr, err := peer.IDFromBytes(msg.From)
			if err != nil {
				fmt.Println("Error occurred when reading message From field...")
				panic(err)
			}

			// This checks if sender address of incoming message is ours. It is need because we get our messages when subscribed to the same topic.
			if addr == node.host.ID() {
				continue
			}
			incomingMessages <- *msg
		}
	}()

	peerChan, err := mdns.InitMDNS(ctx, node.host, topic)
	if err != nil {
		fmt.Println("Failed to init MDNS:", err)
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case newPeer := <-peerChan:
			{
				fmt.Println("\nFound peer:", newPeer, ", add address to peerstore")

				// Adding peer addresses to local peerstore
				node.host.Peerstore().AddAddrs(newPeer.ID, newPeer.Addrs, peerstore.PermanentAddrTTL)
				// Connect to the peer
				if err := node.host.Connect(ctx, newPeer); err != nil {
					fmt.Println("Connection failed:", err)
				}
				handler.Peer(PeerAddr(newPeer))
			}
		case msg := <-incomingMessages:
			{
				err := node.processMessage(msg, topic, handler)
				if err != nil {
					fmt.Println("\nFailed to process message:", err)
				}
			}
		}
	}

	return nil
}

func (node *Node) processMessage(msg pubsub.Message, topic string, handler Handler) error {
	addr, err := peer.IDFromBytes(msg.From)
	if err != nil {
		fmt.Println("Error occurred when reading message From field...")
		return err
	}
	message := api.BaseMessage{}
	err = json.Unmarshal(msg.Data, &message)
	if err != nil {
		return err
	}

	switch message.Flag {

	// Getting regular message
	case api.FLAG_GENERIC_MESSAGE:
		textMessage := TextMessage{
			Body: message.Body,
			From: addr,
		}

		handler.TextMessage(textMessage)

	// Getting topic request, answer topic response
	case api.FLAG_TOPICS_REQUEST:
		respond := &api.GetTopicsRespondMessage{
			BaseMessage: api.BaseMessage{
				Body: "",
				Flag: api.FLAG_TOPICS_RESPONSE,
			},
			Topics: node.pubsub.GetTopics(),
		}
		sendData, err := json.Marshal(respond)
		if err != nil {
			return err
		}
		go func() {
			if err := node.pubsub.Publish(topic, sendData); err != nil {
				fmt.Println("\nFailed to publish:", err)
			}
		}()

	// Getting topic respond, adding topics to `networkTopics`
	case api.FLAG_TOPICS_RESPONSE:
		/*		respond := &api.GetTopicsRespondMessage{}
				err = json.Unmarshal(msg.Data, respond)
				if err != nil {
					return err
				}
				for i := 0; i < len(respond.Topics); i++ {
					h.networkTopics.Add(respond.Topics[i])
				}
		*/
	default:
		return fmt.Errorf("Unknown message type: %#x", message.Flag)
	}
	return nil
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
