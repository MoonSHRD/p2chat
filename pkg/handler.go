package pkg

import (
	"encoding/json"
	"log"
	"sync"

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
	identityMap   map[peer.ID]string
	peerID        peer.ID
	matrixID      string
	PbMutex       sync.Mutex
}

// TextMessage is more end-user model of regular text messages
type TextMessage struct {
	Topic        string `json:"topic"`
	Body         string `json:"body"`
	FromPeerID   string `json:"fromPeerID"`
	FromMatrixID string `json:"fromMatrixID"`
}

func NewHandler(pb *pubsub.PubSub, serviceTopic string, peerID peer.ID, networkTopics *mapset.Set) Handler {
	return Handler{
		pb:            pb,
		serviceTopic:  serviceTopic,
		networkTopics: *networkTopics,
		identityMap:   make(map[peer.ID]string),
		peerID:        peerID,
	}
}

func (h *Handler) HandleIncomingMessage(topic string, msg pubsub.Message, handleTextMessage func(TextMessage), handleMatch func(string, string, string), handleUnmatch func(string, string, string)) {
	fromPeerID, err := peer.IDFromBytes(msg.From)
	if err != nil {
		log.Println("Error occurred when reading message from field...")
		return
	}
	message := &api.BaseMessage{}
	if err = json.Unmarshal(msg.Data, message); err != nil {
		log.Println("Error occurred during unmarshalling the base message data")
		return
	}

	if message.To != "" && message.To != string(h.peerID) {
		return // Drop message, because it is not for us
	}

	switch message.Flag {
	// Getting regular message
	case api.FlagGenericMessage:
		textMessage := TextMessage{
			Topic:        topic,
			Body:         message.Body,
			FromPeerID:   fromPeerID.String(),
			FromMatrixID: message.FromMatrixID,
		}
		handleTextMessage(textMessage)
	// Getting topic request, answer topic response
	case api.FlagTopicsRequest:
		respond := &api.GetTopicsRespondMessage{
			BaseMessage: api.BaseMessage{
				Body:         "",
				Flag:         api.FlagTopicsResponse,
				FromMatrixID: h.matrixID,
				To:           fromPeerID.String(),
			},
			Topics: h.GetTopics(),
		}
		sendData, err := json.Marshal(respond)
		if err != nil {
			log.Println("Error occurred during marshalling the respond from TopicsRequest")
			return
		}
		go func() {
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
		h.sendIdentityResponse(h.serviceTopic, fromPeerID.String())
	// Getting identity respond, mapping Multiaddress/MatrixID
	case api.FlagIdentityResponse:
		h.identityMap[peer.ID(fromPeerID.String())] = message.FromMatrixID
	case api.FlagGreeting:
		handleMatch(topic, fromPeerID.String(), message.FromMatrixID)
		h.sendIdentityResponse(topic, fromPeerID.String())
	case api.FlagGreetingRespond:
		handleMatch(topic, fromPeerID.String(), message.FromMatrixID)
	case api.FlagFarewell:
		handleUnmatch(topic, fromPeerID.String(), message.FromMatrixID)
	default:
		log.Printf("\nUnknown message type: %#x\n", message.Flag)
	}
}

func (h *Handler) sendIdentityResponse(topic string, fromPeerID string) {
	var flag int
	if topic == h.serviceTopic {
		flag = api.FlagIdentityResponse
	} else {
		flag = api.FlagGreetingRespond
	}
	respond := &api.BaseMessage{
		Body:         "",
		Flag:         flag,
		FromMatrixID: h.matrixID,
		To:           fromPeerID,
	}
	sendData, err := json.Marshal(respond)
	if err != nil {
		log.Println("Error occurred during marshalling the respond from IdentityRequest")
		return
	}
	go func() {
		h.PbMutex.Lock()
		h.pb.Publish(topic, sendData)
		h.PbMutex.Unlock()
	}()
}

// Set Matrix ID
func (h *Handler) SetMatrixID(matrixID string) {
	h.matrixID = matrixID
}

// Returns copy of handler's identity map ([peer.ID]=>[matrixID])
func (h *Handler) GetIdentityMap() map[peer.ID]string {
	return h.identityMap
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
func (h *Handler) RequestNetworkTopics() {
	requestTopicsMessage := &api.BaseMessage{
		Body:         "",
		Flag:         api.FlagTopicsRequest,
		To:           "",
		FromMatrixID: h.matrixID,
	}

	h.sendMessageToServiceTopic(requestTopicsMessage)
}

// Requests MatrixID from specific peer
// TODO: refactor with promise
func (h *Handler) RequestPeerIdentity(peerID string) {
	requestPeersIdentity := &api.BaseMessage{
		Body:         "",
		To:           peerID,
		Flag:         api.FlagIdentityRequest,
		FromMatrixID: h.matrixID,
	}

	h.sendMessageToServiceTopic(requestPeersIdentity)
}

// TODO: refactor
func (h *Handler) SendGreetingInTopic(topic string) {
	greetingMessage := &api.BaseMessage{
		Body:         "",
		To:           "",
		Flag:         api.FlagGreeting,
		FromMatrixID: h.matrixID,
	}

	h.sendMessageToTopic(topic, greetingMessage)
}

// TODO: refactor
func (h *Handler) SendFarewellInTopic(topic string) {
	farewellMessage := &api.BaseMessage{
		Body:         "",
		To:           "",
		Flag:         api.FlagFarewell,
		FromMatrixID: h.matrixID,
	}

	h.sendMessageToTopic(topic, farewellMessage)
}

// Sends marshaled message to the service topic
func (h *Handler) sendMessageToServiceTopic(message *api.BaseMessage) {
	h.sendMessageToTopic(h.serviceTopic, message)
}

func (h *Handler) sendMessageToTopic(topic string, message *api.BaseMessage) {
	sendData, err := json.Marshal(message)
	if err != nil {
		log.Println(err.Error())
		return
	}

	go func() {
		h.PbMutex.Lock()
		h.pb.Publish(topic, sendData)
		h.PbMutex.Unlock()
	}()
}
