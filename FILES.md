# cmd/main_test.go  
## Package: p2chat/v2/pkg  
  
### Imports:  
  
```  
context  
crypto/rand  
encoding/json  
fmt  
testing  
time  
  
github.com/MoonSHRD/p2chat/v2/api  
github.com/MoonSHRD/p2chat/v2/pkg  
github.com/libp2p/go-libp2p  
github.com/libp2p/go-libp2p-core/crypto  
github.com/libp2p/go-libp2p-core/host  
github.com/libp2p/go-libp2p-core/peer  
github.com/libp2p/go-libp2p-core/peerstore  
github.com/libp2p/go-libp2p-core/protocol  
pubsub "github.com/libp2p/go-libp2p-pubsub"  
github.com/phayes/freeport  
```  
  
### External Data, Input Sources:  
  
- `numberOfNodes`: Constant representing the number of nodes in the network (3).  
- `serviceTag`: Constant representing the service tag used for communication (moonshard).  
  
### Code Summary:  
  
#### TestCreateHosts:  
  
- Creates multiple mock host objects using `createHost` function.  
- Each host has its own context, private key, and listening port.  
- Appends the created hosts and contexts to respective arrays.  
  
#### TestMDNS:  
  
- Initializes a PubSub instance for each host using `pubsub.NewFloodsubWithProtocols`.  
- Creates a handler for each PubSub instance using `pkg.NewHandler`.  
- Initializes MDNS using `pkg.InitMDNS` and waits for the correct setup of PubSub.  
- Connects each host to the other nodes in the network.  
  
#### TestGetPeers:  
  
- Checks if all nodes are connected to each other by verifying the number of peers returned by `handler.GetPeers`.  
  
#### TestSendMessage:  
  
- Creates a sample message and marshals it into JSON format.  
- Publishes the message to the service topic using `testPubsubs[0].Publish`.  
  
#### TestGetMessage:  
  
- Subscribes to the service topic and receives the published message.  
- Decodes the received message and compares it with the original message.  
  
#### TestCloseHosts:  
  
- Closes all host objects using `host.Close`.  
  
#### End of Output:  
  
  
  
# pkg/handler.go  
## Package: pkg  
  
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
  
# pkg/mdns.go  
## Package: pkg  
  
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
  
  
  
# pkg/mdns_test.go  
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
  
# api/protocol.go  
## Package: api  
  
### Imports:  
  
None  
  
### External Data, Input Sources:  
  
None  
  
### BaseMessage:  
  
The `BaseMessage` struct represents the basic message format of the protocol. It has the following fields:  
  
- `Body`: The message body.  
- `To`: The recipient of the message.  
- `Flag`: An integer representing the message type.  
- `FromMatrixID`: The sender's MatrixID.  
  
### GetTopicsRespondMessage:  
  
The `GetTopicsRespondMessage` struct is used to respond to a request for existing PubSub topics at the network. It inherits from the `BaseMessage` struct and has an additional field:  
  
- `Topics`: A list of strings representing the available PubSub topics.  
  
The `Flag` field for this message type is set to 0x2.  
  
  
  
# cmd/flags.go  
Package: main  
  
Imports:  
- flag  
  
External data, input sources:  
- Command-line flags  
  
## Parsing Flags  
  
This function parses command-line flags and returns a configuration struct. It initializes a new config struct and then uses the flag package to define and parse the following flags:  
  
- `rendezvous`: A string that identifies a group of nodes. This flag is used to connect with friends.  
- `wrapped_host`: The bootstrap node's wrapped host listen address.  
- `pid`: Sets a protocol id for stream headers.  
- `port`: The node's listen port.  
  
The function parses the flags using `flag.Parse()` and returns the populated config struct.  
  
  
  
# cmd/main.go  
## Package: chat-with-mdns  
  
This package provides a simple example for peer discovery using mDNS. mDNS is great when you have multiple peers in a local LAN.  
  
### Imports:  
  
- `bufio`  
- `context`  
- `crypto/rand`  
- `encoding/json`  
- `flag`  
- `fmt`  
- `log`  
- `io`  
- `os`  
- `sync`  
- `time`  
- `github.com/MoonSHRD/p2chat/v2/api`  
- `github.com/MoonSHRD/p2chat/v2/pkg`  
- `github.com/deckarep/golang-set`  
- `github.com/libp2p/go-libp2p`  
- `github.com/libp2p/go-libp2p-core/crypto`  
- `github.com/libp2p/go-libp2p-core/host`  
- `github.com/libp2p/go-libp2p-core/peer`  
- `github.com/libp2p/go-libp2p-core/peerstore`  
- `github.com/libp2p/go-libp2p-core/protocol`  
- `github.com/libp2p/go-libp2p-pubsub`  
- `github.com/multiformats/go-multiaddr`  
  
### External Data, Input Sources:  
  
- Command-line arguments:  
    - `-help`: Display help information.  
    - `-wrapped_host`: Hostname or IP address to listen on.  
    - `-port`: Port to listen on.  
    - `-rendezvous`: Rendezvous string (service tag) to use for peer discovery.  
    - `-pid`: Protocol ID to use for PubSub.  
  
### Code Summary:  
  
1. **Initialization and Setup:**  
    - Create a new RSA key pair for the wrapped host.  
    - Create a new libp2p Host with the specified listen address, identity, and other options.  
    - Initialize a PubSub instance with the specified protocol ID and message signing options.  
    - Create a new Handler instance to handle incoming messages and network topics.  
  
2. **Peer Discovery and Connection:**  
    - Use the provided rendezvous string to discover peers with the same service tag.  
    - Subscribe to the rendezvous topic to receive messages from other peers.  
    - Connect to newly discovered peers and add their addresses to the local peerstore.  
  
3. **Message Handling and Communication:**  
    - Read messages from the PubSub subscription and handle them using the Handler instance.  
    - Write messages to the PubSub subscription using the provided rendezvous string.  
  
4. **Network Topics Management:**  
    - Request network topics from the Handler instance to keep track of all active topics.  
  
5. **Main Loop:**  
    - Continuously listen for incoming messages, new peers, and network topics.  
    - Handle incoming messages and connect to new peers as needed.  
    - Close the host and exit the program when the context is canceled.  
  
6. **Error Handling:**  
    - Handle errors during key generation, host creation, PubSub initialization, peer discovery, and connection.  
  
7. **Logging:**  
    - Log messages, errors, and other relevant information to the console.  
  
8. **Help Information:**  
    - Provide help information when the `-help` flag is set.  
  
This summary provides a high-level overview of the code and its functionality. It covers the main components, data sources, and processes involved in the package.  
  
