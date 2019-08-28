package main

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/json"
	"flag"
	"fmt"
	"log"

	"io"
	"os"
	"sync"

	"time"

	"github.com/MoonSHRD/p2chat/api"
	"github.com/MoonSHRD/p2chat/pkg"
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

	// TODO: Update Readme & checkout and replace better comments




*/

var myself host.Host
var pubSub *pubsub.PubSub

var globalCtx context.Context
var globalCtxCancel context.CancelFunc

var pbMutex sync.Mutex
var networkTopics = mapset.NewSet()
var serviceTopic string

var handler pkg.Handler

// Read messages from subscription (topic)
// NOTE: in this function we are providing subscription object, which means we should subscribe somewhere else before invoke this function
//
func readSub(subscription *pubsub.Subscription, incomingMessagesChan chan pubsub.Message) {
	ctx := globalCtx
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		msg, err := subscription.Next(context.Background())
		if err != nil {
			log.Println("Error reading from buffer")
			panic(err)
		}

		if string(msg.Data) == "" {
			return
		}
		if string(msg.Data) != "\n" {
			addr, err := peer.IDFromBytes(msg.From)
			if err != nil {
				log.Println("Error occurred when reading message From field...")
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
	ctx := globalCtx
	subscription, err := pubSub.Subscribe(topic)
	if err != nil {
		log.Println("Error occurred when subscribing to topic")
		panic(err)
	}
	time.Sleep(3 * time.Second)
	incomingMessages := make(chan pubsub.Message)

	go readSub(subscription, incomingMessages)
	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-incomingMessages:
			{
				handler.HandleIncomingMessage(serviceTopic, msg, func(textMessage pkg.TextMessage) {
					log.Printf("%s \x1b[32m%s\x1b[0m> ", textMessage.From, textMessage.Body)
				})
			}
		}
	}
}

// Write messages to subscription (topic)
// NOTE: we don't need to be subscribed to publish something
func writeTopic(topic string) {
	ctx := globalCtx
	stdReader := bufio.NewReader(os.Stdin)
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		log.Print("> ")
		text, err := stdReader.ReadString('\n')
		if err != nil {

			if err == io.EOF {
				break
			}

			log.Println("Error reading from stdin")
			panic(err)
		}
		message := &api.BaseMessage{
			Body: text,
			Flag: api.FlagGenericMessage,
		}

		sendData, err := json.Marshal(message)
		if err != nil {
			log.Println("Error occurred when marshalling message object")
			continue
		}
		err = pubSub.Publish(topic, sendData)
		if err != nil {
			log.Println("Error occurred when publishing")
			panic(err)
		}
	}
}

func main() {
	help := flag.Bool("help", false, "Display Help")
	cfg := parseFlags()

	if *help {
		log.Printf("Simple example for peer discovery using mDNS. mDNS is great when you have multiple peers in local LAN.")
		log.Printf("Usage: \n   Run './chat-with-mdns'\nor Run './chat-with-mdns -wrapped_host [wrapped_host] -port [port] -rendezvous [string] -pid [proto ID]'\n")

		os.Exit(0)
	}

	log.Printf("[*] Listening on: %s with port: %d\n", cfg.listenHost, cfg.listenPort)

	ctx, ctxCancel := context.WithCancel(context.Background())
	globalCtx = ctx
	globalCtxCancel = ctxCancel

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

	multiaddress := fmt.Sprintf("/ip4/%s/tcp/%v/p2p/%s", cfg.listenHost, cfg.listenPort, host.ID().Pretty())
	log.Printf("\n[*] Your Multiaddress Is: %s\n", multiaddress)

	myself = host

	pb, err := pubsub.NewFloodsubWithProtocols(context.Background(), host, []protocol.ID{protocol.ID(cfg.ProtocolID)}, pubsub.WithMessageSigning(true), pubsub.WithStrictSignatureVerification(true))
	if err != nil {
		log.Println("Error occurred when create PubSub")
		panic(err)
	}

	// Set global PubSub object
	pubSub = pb

	handler = pkg.NewHandler(pb, serviceTopic, multiaddress, &networkTopics)

	// Randezvous string = service tag
	// Disvover all peers with our service (all ms devices)
	peerChan := pkg.InitMDNS(ctx, host, cfg.RendezvousString)

	// NOTE:  here we use Randezvous string as 'topic' by default .. topic != service tag
	subscription, err := pb.Subscribe(cfg.RendezvousString)
	serviceTopic = cfg.RendezvousString
	if err != nil {
		log.Println("Error occurred when subscribing to topic")
		panic(err)
	}

	log.Println("Waiting for correct set up of PubSub...")
	time.Sleep(3 * time.Second)

	incomingMessages := make(chan pubsub.Message)

	go func() {
		writeTopic(cfg.RendezvousString)
		ctxCancel()
	}()
	go readSub(subscription, incomingMessages)
	go getNetworkTopics()

MainLoop:
	for {
		select {
		case <-ctx.Done():
			break MainLoop
		case msg := <-incomingMessages:
			{
				handler.HandleIncomingMessage(serviceTopic, msg, func(textMessage pkg.TextMessage) {
					// Green console colour: 	\x1b[32m
					// Reset console colour: 	\x1b[0m
					log.Printf("%s > \x1b[32m%s\x1b[0m", textMessage.From, textMessage.Body)
					log.Print("> ")
				})
			}
		case newPeer := <-peerChan:
			{
				log.Println("\nFound peer:", newPeer, ", add address to peerstore")

				// Adding peer addresses to local peerstore
				host.Peerstore().AddAddr(newPeer.ID, newPeer.Addrs[0], peerstore.PermanentAddrTTL)
				// Connect to the peer
				if err := host.Connect(ctx, newPeer); err != nil {
					log.Println("Connection failed:", err)
				}
				log.Println("Connected to:", newPeer)
				log.Println("> ")
			}
		}
	}

	if err := host.Close(); err != nil {
		log.Println("\nClosing host failed:", err)
	}
	log.Println("\nBye")
}

func getNetworkTopics() {
	ctx := globalCtx
	handler.RequestNetworkTopics(ctx)
}
