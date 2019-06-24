package p2mobile

import (
//	"bufio"
	"context"
	"crypto/rand"
	"fmt"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/peerstore"
	"github.com/libp2p/go-libp2p-core/protocol"
	"github.com/libp2p/go-libp2p-pubsub"
	"github.com/multiformats/go-multiaddr"

//	"os"
//	"strings"
	"time"

)

//
//
//   structs example in golang
//
// 	 type MobileLibp2p struct {
//  node *host.Host
//  }
//
//	func StartLibp2p() *MobileLibp2p {
//	  host, err := libp2p.New(ctx)
//	  return &MobileLibp2p{
//	    node: &host,
//	  }
//	}
//

//
/*
type StreamApi struct {
	Stream network.Stream
}
*/


/*
// TODO: build & bind tests -- make this firsts

1. try to build
2. try export structures / functions
3. try for android


// TODO: functionality tests
1.	try to connect two devices and test message trasmission between them
2.	try to connect between three different devices onto one topic and test message transmission between them
3.	try getTopics... not sure how it should work

// TODO: get topics lists across the network
Task is to get all topics across network (inside our service 'moonshard').
The pitfall is - we have to subscribe before connection to nodes.
Probably we need to subscribe to moonshard topic, connect to peers, then ask them about what topics are they subscribed to..
Really not sure how it works, need to research


// TODO: next features
1. get topic list across the network
2. implement encryption pubsub.WithMessageSigning(TRUE)
3.




*/

type Config struct {
	RendezvousString string // Unique string to identify group of nodes. Share this with your friends to let them connect with you
	ProtocolID       string // Sets a protocol id for stream headers
	ListenHost       string // The bootstrap node host listen address
	ListenPort       int    // Node listen port
}


var myself host.Host

var Pb *pubsub.PubSub


//======== PubSub related ==========//


// Subscribe to a topic and get messages from it
func SubscribeRead(topic string) string {
	subscription, err := Pb.Subscribe(topic)
	if err != nil {
		fmt.Println("Error occurred when subscribing to topic")
		panic(err)
	}
	time.Sleep(2 * time.Second)
		msg := ReadSub(subscription)
		return msg
}


// this function get new messages from subscribed topic
// working with strings now.. probably be better with data?
func ReadSub(subscription *pubsub.Subscription) string {
	for {
		msg, err := subscription.Next(context.Background())
		if err != nil {
			fmt.Println("Error reading from subscription")
			panic(err)
		}
		// TODO: weird behavior, remove or rework it
		if string(msg.Data) == "" {
			return string(msg.Data)
		}
		if string(msg.Data) != "\n" {
			// Green console colour: 	\x1b[32m
			// Reset console colour: 	\x1b[0m
			addr, err := peer.IDFromBytes(msg.From)
			if err != nil {
				fmt.Println("Error occurred when reading message From field...")
				panic(err)
			}

			if addr == myself.ID() {
				continue
			}
			fmt.Printf("%s \x1b[32m%s\x1b[0m> ", addr,string(msg.Data))
			message := string(msg.Data)
			return message
		}
	}
}



// Publish message into some topic
// working with 'strings' messages. Don't like it
func PublishMessage(topic string, message string)  {
	err := Pb.Publish(topic, []byte(message))
	if err != nil {
		fmt.Println("Error occurred when publishing")
		panic(err)
	}
}




//======Main function========//


// TODO:  why is there types with a pointer? Is it for export?
func Start(rendezvous *string, pid *string, listenHost *string, port *int) {
	cfg := GetConfig(rendezvous, pid, listenHost, port)

	fmt.Printf("[*] Listening on: %s with port: %d\n", cfg.ListenHost, cfg.ListenPort)

	ctx := context.Background()
	r := rand.Reader

	// Creates a new RSA key pair for this host.
	prvKey, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, r)
	if err != nil {
		panic(err)
	}

	// 0.0.0.0 will listen on any interface device.
	sourceMultiAddr, _ := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/%s/tcp/%d", cfg.ListenHost, cfg.ListenPort))

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

	// Set a function as stream handler.
	// This function is called when a peer initiates a connection and starts a stream with this peer. (Handle incoming connections)
//	host.SetStreamHandler(protocol.ID(cfg.ProtocolID), handleStream)

	fmt.Printf("\n[*] Your Multiaddress Is: /ip4/%s/tcp/%v/p2p/%s\n", cfg.ListenHost, cfg.ListenPort, host.ID().Pretty())

	myself = host

	// Initialize pubsub object
	pb, err := pubsub.NewFloodsubWithProtocols(context.Background(), host, []protocol.ID{protocol.ID(cfg.ProtocolID)}, pubsub.WithMessageSigning(false))
	if err != nil {
		fmt.Println("Error occurred when create PubSub")
		panic(err)
	}

	Pb = pb


	peerChan := initMDNS(ctx, host, cfg.RendezvousString)

	peer := <-peerChan // will block untill we discover a peer
	fmt.Println("Found peer:", peer, ", connecting")


	// Adding peer addresses to local peerstore
	host.Peerstore().AddAddr(peer.ID, peer.Addrs[0], peerstore.PermanentAddrTTL)

	// TODO: probably we need somehow to get available topic's list before connect (not sure that we actually can do this before connection.. research needed)



	//Subscription should go BEFORE connections
// NOTE:  here we use Randezvous string as 'topic' by default .. topic != service tag
	subscription, err := pb.Subscribe(cfg.RendezvousString)
	if err != nil {
		fmt.Println("Error occurred when subscribing to topic")
		panic(err)
	}

	// Connect to the peer
	if err := host.Connect(ctx, peer); err != nil {
	fmt.Println("Connection failed:", err)
	}
	fmt.Println("Connected to:", peer)




	fmt.Println("Waiting for correct set up of PubSub...")
	time.Sleep(3 * time.Second)



	//go writeTopic(cfg.RendezvousString)
	go ReadSub(subscription)

	select {} //wait here
}

// TODO: get this part to separate file (flags or whatever). all defaults parameters and their parsing should be done from separate file
func GetConfig(rendezvous *string, pid *string, host *string, port *int) *Config {
	c := &Config{}

	if *rendezvous != "" && rendezvous != nil {
		c.RendezvousString = *rendezvous
	} else {
		c.RendezvousString = "moonshard"
	}

	if *pid != "" && pid != nil {
		c.ProtocolID = *pid
	} else {
		c.ProtocolID = "/chat/1.1.0"
	}

	if *host != "" && host != nil {
		c.ListenHost = *host
	} else {
		c.ListenHost = "0.0.0.0"
	}

	if *port != 0 && port != nil && !(*port < 0) && !(*port > 65535) {
		c.ListenPort = *port
	} else {
		c.ListenPort = 4001
	}

	return c
}
