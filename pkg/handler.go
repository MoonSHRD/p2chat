package pkg

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/MoonSHRD/p2chat/api"
	mapset "github.com/deckarep/golang-set"
	"github.com/libp2p/go-libp2p-core/peer"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
)

// Handler is a network handler, which handle on incoming network events (such as message)
type Handler struct {
	pb            *pubsub.PubSub
	serviceTopic  string
	networkTopics mapset.Set
	PbMutex       sync.Mutex
}

// TextMessage is more end-user model of regular text messages
type TextMessage struct {
	Topic string
	Body  string
	From  peer.ID
}

func NewHandler(pb *pubsub.PubSub, serviceTopic string, networkTopics *mapset.Set) Handler {
	return Handler{
		pb:            pb,
		serviceTopic:  serviceTopic,
		networkTopics: *networkTopics,
	}
}

func (h *Handler) HandleIncomingMessage(topic string, msg pubsub.Message, handleTextMessage func(TextMessage)) {
	addr, err := peer.IDFromBytes(msg.From)
	if err != nil {
		fmt.Println("Error occurred when reading message From field...")
		panic(err)
	}
	message := &api.BaseMessage{}
	err = json.Unmarshal(msg.Data, message)
	if err != nil {
		return
	}
	switch message.Flag {
	// Getting regular message
	case api.FlagGenericMessage:
		textMessage := TextMessage{
			Topic: topic,
			Body:  message.Body,
			From:  addr,
		}

		handleTextMessage(textMessage)

	// Getting topic request, answer topic response
	case api.FlagTopicsRequest:
		respond := &api.GetTopicsRespondMessage{
			BaseMessage: api.BaseMessage{
				Body: "",
				Flag: api.FlagTopicsResponse,
			},
			Topics: h.GetTopics(),
		}
		sendData, err := json.Marshal(respond)
		if err != nil {
			return
		}
		go func() {
			// Lock for blocking "same-time-respond"
			h.PbMutex.Lock()
			h.pb.Publish(h.serviceTopic, sendData)
			h.PbMutex.Unlock()
		}()
	// Getting topic respond, adding topics to `networkTopics`
	case api.FlagTopicsResponse:
		respond := &api.GetTopicsRespondMessage{}
		err = json.Unmarshal(msg.Data, respond)
		if err != nil {
			panic(err)
		}
		for i := 0; i < len(respond.Topics); i++ {
			h.networkTopics.Add(respond.Topics[i])
		}
	default:
		fmt.Printf("\nUnknown message type: %#x\n", message.Flag)
	}
}

// Get list of topics **this** node is subscribed to
func (h *Handler) GetTopics() []string {
	topics := h.pb.GetTopics()
	return topics
}

// Get list of peers subscribed on specific topic
func (h *Handler) GetPeers(topic string) []peer.ID {
	peers := h.pb.ListPeers(topic)
	return peers
}

// Blacklists a peer by its id
func (h *Handler) BlacklistPeer(pid peer.ID) {
	h.pb.BlacklistPeer(pid)
}

// Requesting topics from **other** peers
func (h *Handler) RequestNetworkTopics(ctx context.Context) {

	requestTopicsMessage := &api.BaseMessage{
		Body: "",
		Flag: api.FlagTopicsRequest,
	}
	sendData, err := json.Marshal(requestTopicsMessage)
	if err != nil {
		panic(err)
	}
	t := time.NewTicker(3 * time.Second)
	defer t.Stop()
	for range t.C {
		select {
		case <-ctx.Done():
			return
		default:
		}
		h.PbMutex.Lock()
		h.pb.Publish(h.serviceTopic, sendData)
		h.PbMutex.Unlock()
	}
}
