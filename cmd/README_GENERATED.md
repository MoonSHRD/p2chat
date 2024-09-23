# chat-with-mdns

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

Project package structure:

- cmd/flags.go
- cmd/main.go
- cmd/main_test.go

