package pkg

import (
	"context"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p/p2p/discovery"

	swarmt "github.com/libp2p/go-libp2p-swarm/testing"
	bhost "github.com/libp2p/go-libp2p/p2p/host/basic"
)

func TestInitMDNS(t *testing.T) {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	testHost := bhost.New(swarmt.GenSwarm(t, ctx))

	testMDNS, err := discovery.NewMdnsService(ctx, testHost, time.Hour, "moonshard")
	if err != nil {
		t.Fatal(err)
	}

	n := &discoveryNotifee{}
	n.PeerChan = make(chan peer.AddrInfo)
	testMDNS.RegisterNotifee(n)
}
