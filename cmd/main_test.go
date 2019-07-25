package main

import (
	"context"
	"crypto/rand"
	"fmt"
	"testing"
	"time"

	"github.com/MoonSHRD/p2chat/pkg"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/peerstore"
	"github.com/libp2p/go-libp2p-core/protocol"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/phayes/freeport"
)

const (
	numberOfNodes = 3
	serviceTag    = "moonshard"
)

var (
	testHosts    []host.Host
	testContexts []context.Context
	testHandlers []pkg.Handler
	peerChan     chan peer.AddrInfo
)

// Creates mock host object
func createHost() (context.Context, host.Host, error) {
	ctx, _ /* cancel */ := context.WithCancel(context.Background())
	// defer cancel()

	prvKey, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, rand.Reader)
	if err != nil {
		return nil, nil, err
	}

	port, err := freeport.GetFreePort()
	if err != nil {
		return nil, nil, err
	}

	host, err := libp2p.New(
		ctx,
		libp2p.Identity(prvKey),
		libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/0.0.0.0/tcp/%v", port)),
	)
	if err != nil {
		return nil, nil, err
	}

	return ctx, host, nil
}

func TestCreateHosts(t *testing.T) {
	for i := 0; i < numberOfNodes; i++ {
		tempCtx, tempHost, err := createHost()
		if err != nil {
			t.Fatal(err)
		}

		testHosts = append(testHosts, tempHost)
		testContexts = append(testContexts, tempCtx)
	}
}

func TestMDNS(t *testing.T) {
	for i := 0; i < numberOfNodes; i++ {
		pb, err := pubsub.NewFloodsubWithProtocols(context.Background(), testHosts[i], []protocol.ID{protocol.ID("/moonshard/1.0.0")}, pubsub.WithMessageSigning(true), pubsub.WithStrictSignatureVerification(true))
		if err != nil {
			t.Fatal(err)
		}

		testHandlers = append(testHandlers, pkg.NewHandler(pb, serviceTag, &networkTopics))

		peerChan = pkg.InitMDNS(testContexts[i], testHosts[i], serviceTag)

		_, err = pb.Subscribe(serviceTag)
		if err != nil {
			t.Fatal(err)
		}

		fmt.Println("Waiting for correct set up of PubSub...")
		time.Sleep(3 * time.Second)

		for j := 0; j < i; j++ {
			select {
			case peer := <-peerChan:
				testHosts[i].Peerstore().AddAddr(peer.ID, peer.Addrs[0], peerstore.PermanentAddrTTL)

				if err := testHosts[i].Connect(testContexts[i], peer); err != nil {
					t.Fatal(err)
				}
			default:
			}
		}
	}
}

// Checks whether all nodes are connected to each other
func TestGetPeers(t *testing.T) {
	for i := range testHandlers {
		if len(testHandlers[i].GetPeers(serviceTag)) != numberOfNodes-1 {
			t.Fatal("Not all nodes are connected to each other.")
		}
	}
}

func TestCloseHosts(t *testing.T) {
	for _, host := range testHosts {
		if err := host.Close(); err != nil {
			t.Fatal(fmt.Sprintf("Failed when closing host %v", host.ID()))
		}
	}
}
