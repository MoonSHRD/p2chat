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

	api "github.com/MoonSHRD/p2chat/api"
	internal "github.com/MoonSHRD/p2chat/internal"
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
	2. Update handlers in p2mobile (getters / setters) etc.
	3. Update export types in p2mobile
	4. Add exposure functionality with topics (get topics list etc.)
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
func newTopic(topic string) {
	subscription, err := pubSub.Subscribe(topic)
	if err != nil {
		fmt.Println("Error occurred when subscribing to topic")
		panic(err)
	}
	time.Sleep(3 * time.Second)
	incomingMessages := make(chan pubsub.Message)
	go readSub(subscription, incomingMessages)
	for {
		select {
		case msg := <-incomingMessages:
			{
				handleIncomingMessage(msg)
			}
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
		message := &api.BaseMessage{
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

	pb, err := pubsub.NewFloodsubWithProtocols(context.Background(), host, []protocol.ID{protocol.ID(cfg.ProtocolID)}, pubsub.WithMessageSigning(true), pubsub.WithStrictSignatureVerification(true))
	if err != nil {
		fmt.Println("Error occurred when create PubSub")
		panic(err)
	}

	pubSub = pb

	// Randezvous string = service tag
	// Disvover all peers with our service (all ms devices)
	peerChan := internal.InitMDNS(ctx, host, cfg.RendezvousString)

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
				handleIncomingMessage(msg)
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
		getTopicsMessage := &api.BaseMessage{
			Body: "",
			Flag: 0x1,
		}
		sendData, err := json.Marshal(getTopicsMessage)
		if err != nil {
			continue
		}
		time.Sleep(2 * time.Second)
		pubSub.Publish(serviceTopic, sendData)
		time.Sleep(3 * time.Second)
	}
}

func handleIncomingMessage(msg pubsub.Message) {
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
			Topics: getTopics(),
		}
		sendData, err := json.Marshal(ack)
		if err != nil {
			return
		}
		go func() {
			time.Sleep(1 * time.Second)
			pubSub.Publish(serviceTopic, sendData)
		}()
	} else if message.Flag == 0x2 {
		ack := &api.GetTopicsAckMessage{}
		err = json.Unmarshal(msg.Data, ack)
		if err != nil {
			return
		}
		for i := 0; i < len(ack.Topics); i++ {
			networkTopics.Add(ack.Topics[i])
		}
	}
}
