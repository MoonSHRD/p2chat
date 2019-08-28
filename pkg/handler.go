package pkg

import (
	"context"
	"encoding/json"
	"log"
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
	identityMap   map[string]string
	multiaddress  string
	PbMutex       sync.Mutex
}

// TextMessage is more end-user model of regular text messages
type TextMessage struct {
	Topic string
	Body  string
	From  peer.ID
}

func NewHandler(pb *pubsub.PubSub, serviceTopic, multiaddress string, networkTopics *mapset.Set) Handler {
	return Handler{
		pb:            pb,
		serviceTopic:  serviceTopic,
		networkTopics: *networkTopics,
		identityMap:   make(map[string]string),
		multiaddress:  multiaddress,
	}
}

func (h *Handler) HandleIncomingMessage(topic string, msg pubsub.Message, handleTextMessage func(TextMessage)) {
	addr, err := peer.IDFromBytes(msg.From)
	if err != nil {
		log.Println("Error occurred when reading message from field...")
		return
	}
	message := &api.BaseMessage{}
	if err = json.Unmarshal(msg.Data, message); err != nil {
		log.Println("Error occurred during unmarshalling the base message data")
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
			log.Println("Error occurred during marshalling the respond from TopicsRequest")
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
		if err = json.Unmarshal(msg.Data, respond); err != nil {
			log.Println("Error occurred during unmarshalling the message data from TopicsResponse")
			return
		}
		for i := 0; i < len(respond.Topics); i++ {
			h.networkTopics.Add(respond.Topics[i])
		}
	// Getting identity request, answer identity response
	case api.FlagIdentityRequest:
		respond := &api.GetIdentityRespondMessage{
			BaseMessage: api.BaseMessage{
				Body: "",
				Flag: api.FlagIdentityResponse,
			},
			Multiaddress: h.multiaddress,
			MatrixID:     "",
		}
		sendData, err := json.Marshal(respond)
		if err != nil {
			log.Println("Error occurred during marshalling the respond from IdentityRequest")
			return
		}
		go func() {
			h.PbMutex.Lock()
			h.pb.Publish(h.serviceTopic, sendData)
			h.PbMutex.Unlock()
		}()
	// Getting identity respond, mapping Multiaddress/MatrixID
	case api.FlagIdentityResponse:
		respond := &api.GetIdentityRespondMessage{}
		if err := json.Unmarshal(msg.Data, respond); err != nil {
			log.Println("Error occurred during unmarshalling the message data from IdentityResponse")
			return
		}
		h.identityMap[respond.Multiaddress] = respond.MatrixID
	default:
		log.Printf("\nUnknown message type: %#x\n", message.Flag)
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

	h.sendMessageToServiceTopic(ctx, requestTopicsMessage)
}

// Requests MatrixID from other peers
func (h *Handler) RequestPeersIdentity(ctx context.Context) {
	requestPeersIdentity := &api.BaseMessage{
		Body: "",
		Flag: api.FlagIdentityRequest,
	}

	h.sendMessageToServiceTopic(ctx, requestPeersIdentity)
}

// Sends marshaled message to the service topic
func (h *Handler) sendMessageToServiceTopic(ctx context.Context, message *api.BaseMessage) {
	sendData, err := json.Marshal(message)
	if err != nil {
		log.Println(err.Error())
		return
	}

	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
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
