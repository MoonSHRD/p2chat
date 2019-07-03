package main

import (
	"context"
	"crypto/rand"
	"flag"
	"fmt"
	"os"

	"github.com/MoonSHRD/p2chat/internal/cli"
	"github.com/MoonSHRD/p2chat/pkg/p2chat"

	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/multiformats/go-multiaddr"
)

/*

	// TODO: Update Readme & checkout and replace better comments




*/

func main() {
	help := flag.Bool("help", false, "Display Help")
	cfg := parseFlags()

	if *help {
		fmt.Printf("Simple example for peer discovery using mDNS. mDNS is great when you have multiple peers in local LAN.")
		fmt.Printf("Usage: \n   Run './chat-with-mdns'\nor Run './chat-with-mdns -wrapped_host [wrapped_host] -port [port] -rendezvous [string] -pid [proto ID]'\n")

		os.Exit(0)
	}

	fmt.Printf("[*] Listening on: %s with port: %d\n", cfg.listenHost, cfg.listenPort)

	ctx, ctxCancel := context.WithCancel(context.Background())

	// Creates a new RSA key pair for this wrapped_host.
	prvKey, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, rand.Reader)
	if err != nil {
		panic(err)
	}

	// 0.0.0.0 will listen on any interface device.
	sourceMultiAddr, err := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/%s/tcp/%d", cfg.listenHost, cfg.listenPort))
	if err != nil {
		panic(err)
	}

	chatNode, err := p2chat.NewNode(
		ctx,
		sourceMultiAddr,
		prvKey,
		cfg.ProtocolID,
	)
	if err != nil {
		panic(err)
	}

	fmt.Printf("\n[*] Your Multiaddress Is: /ip4/%s/tcp/%v/p2p/%s\n", cfg.listenHost, cfg.listenPort, chatNode.HostID())

	cliHandler := cli.NewHandler(chatNode, cfg.RendezvousString)

	go func() {
		err := chatNode.Subscribe(ctx, cfg.RendezvousString, cliHandler)
		if err != nil {
			fmt.Println("\nFailed to subscribe:", err)
		}
		ctxCancel()
	}()

	cliHandler.ReadStdin(ctx)
	ctxCancel()

	if err := chatNode.Close(); err != nil {
		fmt.Println("\nClosing host failed:", err)
	}
	fmt.Println("\nBye")
}
