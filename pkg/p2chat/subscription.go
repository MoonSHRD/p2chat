package p2chat

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/MoonSHRD/p2chat/api"
	"github.com/MoonSHRD/p2chat/pkg/mdns"

	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/peerstore"
	"github.com/libp2p/go-libp2p-pubsub"
)

type subscription struct {
	node    *Node
	topic   string
	handler Handler
}

func newSubscription(ctx context.Context, node *Node, topic string, handler Handler) error {
	s := subscription{
		node:    node,
		topic:   topic,
		handler: handler,
	}
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
				log.Printf("Error reading from buffer: %s", err)
				continue
			}

			if len(msg.Data) == 0 {
				continue
			}
			addr, err := peer.IDFromBytes(msg.From)
			if err != nil {
				log.Printf("Error occurred when reading message From field: %s", err)
				continue
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
		log.Printf("Failed to init MDNS: %s", err)
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case newPeer := <-peerChan:
			{
				log.Printf("Found peer: %s, add address to peerstore", newPeer)

				// Adding peer addresses to local peerstore
				node.host.Peerstore().AddAddrs(newPeer.ID, newPeer.Addrs, peerstore.PermanentAddrTTL)
				// Connect to the peer
				if err := node.host.Connect(ctx, newPeer); err != nil {
					log.Printf("Connection failed: %s", err)
				}
				handler.Peer(PeerAddr(newPeer))
			}
		case msg := <-incomingMessages:
			{
				err := s.processMessage(msg)
				if err != nil {
					log.Printf("Failed to process message: %s", err)
				}
			}
		}
	}

	return nil
}

func (s *subscription) processMessage(msg pubsub.Message) error {
	addr, err := peer.IDFromBytes(msg.From)
	if err != nil {
		log.Printf("Error occurred when reading message From field: %s", err)
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

		s.handler.TextMessage(textMessage)

	// Getting topic request, answer topic response
	case api.FLAG_TOPICS_REQUEST:
		respond := &api.GetTopicsRespondMessage{
			BaseMessage: api.BaseMessage{
				Body: "",
				Flag: api.FLAG_TOPICS_RESPONSE,
			},
			Topics: s.node.pubsub.GetTopics(),
		}
		sendData, err := json.Marshal(respond)
		if err != nil {
			return err
		}
		go func() {
			if err := s.node.pubsub.Publish(s.topic, sendData); err != nil {
				log.Printf("Failed to publish: %s", err)
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
