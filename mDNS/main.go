package main

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	mapset "github.com/deckarep/golang-set"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/peerstore"
	"github.com/libp2p/go-libp2p-core/protocol"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/multiformats/go-multiaddr"
)

/*

	// TODO:
	0.
	1.
	2. Update handlers in p2mobile (getters / setters) e.t.c.
	3. Update export types in p2mobile
	4. Add exposure functionality with topics (get topics list e.t.c.)
	5. Add message signing and work with identity (pubsub.WithMessageSigning(TRUE)), try topic validators (??)


	//------------------------------

	// TODO: -- in this file --
	1. newTopic function
	2. getTopic list (probably also getTopics across network?)

*/

var myself host.Host
var pubSub *pubsub.PubSub
var networkTopics = mapset.NewSet()
var serviceTopic string

/*
	The basic message format of our protocol
	Flags:
		- 0x0: Generic message
		- 0x1: Request to get existing PubSub topics on the network
		- 0x2: Answer to the request for topics (ack)
*/
type BaseMessage struct {
	Body string `json:"body"`
	Flag int    `json:"flag"`
}

/*
	The format of the message to answer of request for topics
	Flag: 0x2
*/
type GetTopicsAckMessage struct {
	BaseMessage
	Topics []string `json:"topics"`
}

// Read messages from subscription (topic)
// NOTE: in this function we are providing subscription object, which means we should subscribe somewhere else before invoke this function
// it could be replaced by getting global Pb object..?
func readSub(subscription *pubsub.Subscription, incomingMessagesChan chan pubsub.Message) {
	for {
		msg, err := subscription.Next(context.Background())
		if err != nil {
			fmt.Println("Error reading from buffer")
			panic(err)
		}

		if string(msg.Data) == "" {
			return
		}
		if string(msg.Data) != "\n" {
			addr, err := peer.IDFromBytes(msg.From)
			if err != nil {
				fmt.Println("Error occurred when reading message From field...")
				panic(err)
			}

			// This checks if sender address of incoming message is ours. It is need because we get our messages when subscribed to the same topic.
			if addr == myself.ID() {
				continue
			}
			incomingMessagesChan <- *msg
		}

	}
}

// Subscribes to a topic and then get messages ..
func subscribeRead(topic string) {
	subscription, err := pubSub.Subscribe(topic)
	if err != nil {
		fmt.Println("Error occurred when subscribing to topic")
		panic(err)
	}
	time.Sleep(2 * time.Second)
	incomingMessages := make(chan pubsub.Message)
	readSub(subscription, incomingMessages)
	select {
	case msg := <-incomingMessages:
		{
			addr, err := peer.IDFromBytes(msg.From)
			if err != nil {
				fmt.Println("Error occurred when reading message From field...")
				panic(err)
			}
			fmt.Printf("\x1b[32m%s\x1b[0m> %s", addr, string(msg.Data))
		}
	}
}

// Get list of topics this node is subscribed to
func getTopics() []string {
	topics := pubSub.GetTopics()
	return topics
}

// Get list of peers we connected to a specified topic
func getTopicMembers(topic string) []peer.ID {
	members := pubSub.ListPeers(topic)
	return members
}

// Initialize new chat with given topic string
// this node will subscribe to a new messages and discovery for our topic and publish a hello message
func newTopic(topic string) {
	sendData := string("hello") // TODO: should be replaced with standardized protocol message
	// probably don't need to subscribe
	subscription, err := pubSub.Subscribe(topic)
	if err != nil {
		fmt.Println("Error occurred when subscribing to topic")
		panic(err)
	}
	fmt.Println("subscription:", subscription)
	time.Sleep(2 * time.Second)
	err = pubSub.Publish(topic, []byte(sendData))
	if err != nil {
		fmt.Println("Error occurred when publishing")
		panic(err)
	}
}

// Write messages to subscription (topic)
// NOTE: we don't need to be subscribed to publish something
func writeTopic(topic string) {
	stdReader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("> ")
		text, err := stdReader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading from stdin")
			panic(err)
		}
		message := &BaseMessage{
			Body: text,
			Flag: 0x0,
		}
		sendData, err := json.Marshal(message)
		if err != nil {
			fmt.Println("Error occurred when marshalling message object")
			continue
		}
		err = pubSub.Publish(topic, sendData)
		if err != nil {
			fmt.Println("Error occurred when publishing")
			panic(err)
		}
	}
}

func main() {
	help := flag.Bool("help", false, "Display Help")
	cfg := parseFlags()

	if *help {
		fmt.Printf("Simple example for peer discovery using mDNS. mDNS is great when you have multiple peers in local LAN.")
		fmt.Printf("Usage: \n   Run './chat-with-mdns'\nor Run './chat-with-mdns -wrapped_host [wrapped_host] -port [port] -rendezvous [string] -pid [proto ID]'\n")

		os.Exit(0)
	}

	fmt.Printf("[*] Listening on: %s with port: %d\n", cfg.listenHost, cfg.listenPort)

	ctx := context.Background()
	r := rand.Reader

	// Creates a new RSA key pair for this wrapped_host.
	prvKey, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, r)
	if err != nil {
		panic(err)
	}

	// 0.0.0.0 will listen on any interface device.
	sourceMultiAddr, _ := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/%s/tcp/%d", cfg.listenHost, cfg.listenPort))

	// libp2p.New constructs a new libp2p Host.
	// Other options can be added here.
	host, err := libp2p.New(
		ctx,
		libp2p.ListenAddrs(sourceMultiAddr),
		libp2p.Identity(prvKey),
	)

	if err != nil {
		panic(err)
	}

	fmt.Printf("\n[*] Your Multiaddress Is: /ip4/%s/tcp/%v/p2p/%s\n", cfg.listenHost, cfg.listenPort, host.ID().Pretty())

	myself = host

	pb, err := pubsub.NewFloodsubWithProtocols(context.Background(), host, []protocol.ID{protocol.ID(cfg.ProtocolID)}, pubsub.WithMessageSigning(false))
	if err != nil {
		fmt.Println("Error occurred when create PubSub")
		panic(err)
	}

	pubSub = pb

	// Randezvous string = service tag
	// Disvover all peers with our service (all ms devices)
	peerChan := initMDNS(ctx, host, cfg.RendezvousString)

	// NOTE:  here we use Randezvous string as 'topic' by default .. topic != service tag
	subscription, err := pb.Subscribe(cfg.RendezvousString)
	serviceTopic = cfg.RendezvousString
	if err != nil {
		fmt.Println("Error occurred when subscribing to topic")
		panic(err)
	}

	fmt.Println("Waiting for correct set up of PubSub...")
	time.Sleep(3 * time.Second)

	incomingMessages := make(chan pubsub.Message)

	go writeTopic(cfg.RendezvousString)
	go readSub(subscription, incomingMessages)
	go getNetworkTopics()

	for {
		select {
		case msg := <-incomingMessages:
			{
				addr, err := peer.IDFromBytes(msg.From)
				if err != nil {
					fmt.Println("Error occurred when reading message From field...")
					panic(err)
				}
				message := &BaseMessage{}
				err = json.Unmarshal(msg.Data, message)
				if err != nil {
					continue
				}
				if message.Flag == 0x0 {
					// Green console colour: 	\x1b[32m
					// Reset console colour: 	\x1b[0m
					fmt.Printf("%s \x1b[32m%s\x1b[0m> ", addr, message.Body)
				} else if message.Flag == 0x1 {
					ack := &GetTopicsAckMessage{
						BaseMessage: BaseMessage{
							Body: "",
							Flag: 0x2,
						},
						Topics: getTopics(),
					}
					sendData, err := json.Marshal(ack)
					if err != nil {
						continue
					}
					pb.Publish(cfg.RendezvousString, sendData)
				} else if message.Flag == 0x2 {
					ack := &GetTopicsAckMessage{}
					err = json.Unmarshal(msg.Data, ack)
					if err != nil {
						continue
					}
					for i := 0; i < len(ack.Topics); i++ {
						networkTopics.Add(ack.Topics[i])
					}
				}
			}
		case newPeer := <-peerChan:
			{
				fmt.Println("\nFound peer:", newPeer, ", add address to peerstore")

				// Adding peer addresses to local peerstore
				host.Peerstore().AddAddr(newPeer.ID, newPeer.Addrs[0], peerstore.PermanentAddrTTL)
				// Connect to the peer
				if err := host.Connect(ctx, newPeer); err != nil {
					fmt.Println("Connection failed:", err)
				}
				fmt.Println("\nConnected to:", newPeer)
			}
		}
	}
}

func getNetworkTopics() {
	for {
		getTopicsMessage := &BaseMessage{
			Body: "",
			Flag: 0x1,
		}
		sendData, err := json.Marshal(getTopicsMessage)
		if err != nil {
			continue
		}
		pubSub.Publish(serviceTopic, sendData)
		time.Sleep(3 * time.Second)
	}
}
