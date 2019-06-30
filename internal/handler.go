package internal

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/MoonSHRD/p2chat/api"
	mapset "github.com/deckarep/golang-set"
	peer "github.com/libp2p/go-libp2p-peer"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
)

type Handler struct {
	pb            *pubsub.PubSub
	serviceTopic  string
	networkTopics mapset.Set
}

func NewHandler(pb *pubsub.PubSub, serviceTopic string, networkTopics *mapset.Set) Handler {
	return Handler{
		pb:            pb,
		serviceTopic:  serviceTopic,
		networkTopics: *networkTopics,
	}
}

func (h *Handler) HandleIncomingMessage(msg pubsub.Message) {
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
	if message.Flag == 0x0 {
		// Green console colour: 	\x1b[32m
		// Reset console colour: 	\x1b[0m
		fmt.Printf("%s \x1b[32m%s\x1b[0m> ", addr, message.Body)
	} else if message.Flag == 0x1 {
		ack := &api.GetTopicsAckMessage{
			BaseMessage: api.BaseMessage{
				Body: "",
				Flag: 0x2,
			},
			Topics: h.getTopics(),
		}
		sendData, err := json.Marshal(ack)
		if err != nil {
			return
		}
		go func() {
			time.Sleep(1 * time.Second)
			h.pb.Publish(h.serviceTopic, sendData)
		}()
	} else if message.Flag == 0x2 {
		ack := &api.GetTopicsAckMessage{}
		err = json.Unmarshal(msg.Data, ack)
		if err != nil {
			return
		}
		for i := 0; i < len(ack.Topics); i++ {
			h.networkTopics.Add(ack.Topics[i])
		}
	}
}

// Get list of topics this node is subscribed to
func (h *Handler) getTopics() []string {
	topics := h.pb.GetTopics()
	return topics
}
