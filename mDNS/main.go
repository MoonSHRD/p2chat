package main

import (
	"bufio"
	"context"
	"crypto/rand"
	"flag"
	"fmt"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/peerstore"
	"github.com/libp2p/go-libp2p-core/protocol"
	"github.com/libp2p/go-libp2p-pubsub"
	"github.com/multiformats/go-multiaddr"
	"os"
	"time"
)

/*

	// TODO:
	1. Update p2mobile
	2. Update handlers in p2mobile (getters / setters) e.t.c.
	3. Update export types in p2mobile
	4. Add exposure functionality with topics (get topics list e.t.c.)
	5. Add message signing and work with identity (pubsub.WithMessageSigning(TRUE)), try topic validators (??)

*/

var myself host.Host
var Pb *pubsub.PubSub

func readData(subscription *pubsub.Subscription) {
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
		}

	}
}

func writeData(topic string) {
	stdReader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("> ")
		sendData, err := stdReader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading from stdin")
			panic(err)
		}

		err = Pb.Publish(topic, []byte(sendData))
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

	Pb = pb


// Randezvous string = service tag
	// Disvover all peers with our service (all ms devices)
	peerChan := initMDNS(ctx, host, cfg.RendezvousString)

	peer := <-peerChan // will block untill we discover a peer
	fmt.Println("Found peer:", peer, ", add address to peerstore")

	// Adding peer addresses to local peerstore
	host.Peerstore().AddAddr(peer.ID, peer.Addrs[0], peerstore.PermanentAddrTTL)



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

	go writeData(cfg.RendezvousString)
	go readData(subscription)

	select {} //wait here
}
