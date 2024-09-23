## Package: pkg

### Imports:

- `context`
- `testing`
- `time`
- `github.com/libp2p/go-libp2p-core/peer`
- `github.com/libp2p/go-libp2p/p2p/discovery`
- `github.com/libp2p/go-libp2p-swarm/testing`
- `github.com/libp2p/go-libp2p/p2p/host/basic`

### External Data, Input Sources:

- `context.Background()`
- `time.Hour`
- `moonshard`

### TestInitMDNS:

This function tests the initialization of an MDNS service. It creates a new Libp2p host using `bhost.New` and a swarm using `swarmt.GenSwarm`. Then, it creates a new MDNS service using `discovery.NewMdnsService` with the host, a timeout of one hour, and the service name "moonshard".

A new discovery notifee is created and registered with the MDNS service. The notifee has a channel for receiving peer address information. The test function then proceeds to verify that the MDNS service is initialized correctly.