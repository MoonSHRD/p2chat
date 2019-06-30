package internal

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"
	"context"

	"github.com/MoonSHRD/p2chat/api"
	mapset "github.com/deckarep/golang-set"
	peer "github.com/libp2p/go-libp2p-peer"
	pubsub "github.com/libp2p/go-libp2p-pubsub"

)

type Handler struct {
	pb            *pubsub.PubSub
	serviceTopic  string
	networkTopics mapset.Set
	pbMutex				sync.Mutex
}

func NewHandler(pb *pubsub.PubSub, serviceTopic string, networkTopics *mapset.Set) Handler {
	return Handler{
		pb:            pb,
		serviceTopic:  serviceTopic,
		networkTopics: *networkTopics,
	}
}

func (h *Handler) HandleIncomingMessage(msg pubsub.Message) {
//	var pbMutex sync.Mutex
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
	case api.FLAG_GENERIC_MESSAGE:
		// Green console colour: 	\x1b[32m
		// Reset console colour: 	\x1b[0m
		fmt.Printf("%s \x1b[32m%s\x1b[0m> ", addr, message.Body)

	// Getting topic request, answer topic response
	case api.FLAG_TOPICS_REQUEST:
		respond := &api.GetTopicsRespondMessage{
			BaseMessage: api.BaseMessage{
				Body: "",
				Flag: api.FLAG_TOPICS_RESPONSE,
			},
			Topics: h.getTopics(),
		}
		sendData, err := json.Marshal(respond)
		if err != nil {
			return
		}
		go func() {
			// Lock for blocking "same-time-respond"
			h.pbMutex.Lock()
			h.pb.Publish(h.serviceTopic, sendData)
			h.pbMutex.Unlock()
		}()
	// Getting topic respond, adding topics to `networkTopics`
	case api.FLAG_TOPICS_RESPONSE:
		respond := &api.GetTopicsRespondMessage{}
		err = json.Unmarshal(msg.Data, respond)
		if err != nil {
			return
		}
		for i := 0; i < len(respond.Topics); i++ {
			h.networkTopics.Add(respond.Topics[i])
		}
	default:
		fmt.Printf("\nUnknown message type: %#x\n", message.Flag)
	}
}

// Get list of topics **this** node is subscribed to
func (h *Handler) getTopics() []string {
	topics := h.pb.GetTopics()
	return topics
}

// Requesting topics from **other** peers
func (h *Handler) RequestNetworkTopics(ctx context.Context) {

	requestTopicsMessage := &api.BaseMessage{
		Body: "",
		Flag: api.FLAG_TOPICS_REQUEST,
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
		h.pbMutex.Lock()
		h.pb.Publish(h.serviceTopic, sendData)
		h.pbMutex.Unlock()
	}
}
