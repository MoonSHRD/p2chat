package main

import (
	"context"
	"fmt"
	"time"

	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/peerstore"
	"github.com/libp2p/go-libp2p/p2p/discovery"
)

type discoveryNotifee struct {
	host     host.Host
}

//interface to be called when new  peer is found
func (n *discoveryNotifee) HandlePeerFound(pi peer.AddrInfo) {
	n.host.Peerstore().AddAddr(pi.ID, pi.Addrs[0], peerstore.PermanentAddrTTL)
	fmt.Println("Peer found!")
}

//Initialize the MDNS service
func initMDNS(ctx context.Context, host *host.Host, rendezvous string) {
	// An hour might be a long long period in practical applications. But this is fine for us
	ser, err := discovery.NewMdnsService(ctx, *host, time.Hour, rendezvous)
	if err != nil {
		panic(err)
	}

	// Register with service so that we get notified about peer discovery
	n := &discoveryNotifee{
		host: *host,
	}

	ser.RegisterNotifee(n)
}
