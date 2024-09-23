# Package: pkg

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

### Imports:

- encoding/json
- log
- sync
- github.com/MoonSHRD/p2chat/v2/api
- github.com/deckarep/golang-set
- github.com/libp2p/go-libp2p-core/peer
- github.com/libp2p/go-libp2p-pubsub

### External Data, Input Sources:

- PubSub instance (pb)
- Service topic (serviceTopic)
- Network topics (networkTopics)
- Identity map (identityMap)
- Peer ID (peerID)
- Matrix ID (matrixID)

### Code Summary:

#### Handler struct:

- The Handler struct is responsible for handling incoming network events, such as messages.
- It has a PubSub instance (pb), a service topic (serviceTopic), a set of network topics (networkTopics), an identity map (identityMap), a peer ID (peerID), and a matrix ID (matrixID).

#### TextMessage struct:

- The TextMessage struct represents a regular text message with fields for topic, body, fromPeerID, and fromMatrixID.

#### NewHandler function:

- Creates a new Handler instance with the given PubSub instance, service topic, peer ID, and network topics.

#### HandleIncomingMessage function:

- Handles incoming messages by extracting the message type, peer ID, and matrix ID.
- Based on the message type, it performs different actions, such as sending a response, adding topics to the network topics set, or mapping Multiaddress/MatrixID.

#### GetTopics function:

- Returns a list of topics that the handler is subscribed to.

#### GetPeers function:

- Returns a list of peers subscribed to a specific topic.

#### BlacklistPeer function:

- Blacklists a peer by its ID.

#### RequestNetworkTopics function:

- Requests topics from other peers.

#### RequestPeerIdentity function:

- Requests the MatrixID of a specific peer.

#### SendGreetingInTopic and SendFarewellInTopic functions:

- Sends greeting and farewell messages to a specific topic.

#### sendMessageToServiceTopic and sendMessageToTopic functions:

- Sends marshaled messages to the service topic or a specific topic.

#### SetMatrixID function:

- Sets the Matrix ID for the handler.

#### GetIdentityMap function:

- Returns a copy of the handler's identity map.

### Imports:

- context
- log
- time
- github.com/libp2p/go-libp2p-core/host
- github.com/libp2p/go-libp2p-core/peer
- github.com/libp2p/go-libp2p/p2p/discovery

### External Data, Input Sources:

- context.Context
- host.Host
- rendezvous string

### Code Summary:

#### discoveryNotifee struct:

This struct is used to receive notifications about newly discovered peers. It has a channel called PeerChan that will receive peer.AddrInfo when a new peer is found.

#### HandlePeerFound function:

This function is called when a new peer is found. It takes a peer.AddrInfo as input and sends it to the PeerChan channel.

#### InitMDNS function:

This function initializes the MDNS service and returns a channel that will receive notifications about newly discovered peers. It takes a context.Context, host.Host, and rendezvous string as input.

1. It creates a new MDNS service using the provided context, host, and a one-hour timeout.
2. It registers a discoveryNotifee instance with the service to receive notifications about new peers.
3. It returns a channel that will receive peer.AddrInfo when a new peer is found.



