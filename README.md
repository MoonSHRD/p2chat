# P2Chat
P2Chat - is a core local messenger library, which based on Libp2p stack.

P2Chat basicaly supports discovery through **mDNS** service and support messaging via **PubSub**

It supports following features:
- devices autodiscovery by `Rendezvous string`
- topic list exchanging between peers
- autoconnect group chats by `PubSub`
- default signing and validating messages (crypto)
- crossplatform

## How it works?

```
  cmd/main.go - main logic
  pkg/mdns.go - mdns logic
  api/protocol.go - protocol logic (message struct and handle)
```

### Example work scenario

### Step 1 - establising a network

As first step we want to discover every our peer in network and connect to them.

First thing we need to do is parse configuration of ourselves:

` sourceMultiAddr, _ := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/%s/tcp/%d", cfg.listenHost, cfg.listenPort)) `

multiaddress is universal address of our host, based on our IP address and host, it is basic IPFS identity
if IP is not known, then we could use `0.0.0.0` as a default wildcard for _ourselves_



Then we create `libp2p` host object:

`host, err := libp2p.New(
		ctx,
		libp2p.ListenAddrs(sourceMultiAddr),
		libp2p.Identity(prvKey),
	)
`

After we have been created a host, we could start __peerdiscovery__ using __mDNS__
This is not the only way to peerdiscovery mechanism; alternatively we could use DHT or rendez-vous point as we wish.

`peerChan := pkg.InitMDNS(ctx, host, cfg.RendezvousString) `

Each time we discover a new peer in serviceTopic, we add it to a local peerstore and _connect_ to it:

```
case newPeer := <-peerChan:
			{
				fmt.Println("\nFound peer:", newPeer, ", add address to peerstore")

				// Adding peer addresses to local peerstore
				host.Peerstore().AddAddr(newPeer.ID, newPeer.Addrs[0], peerstore.PermanentAddrTTL)
				// Connect to the peer
				if err := host.Connect(ctx, newPeer); err != nil {
					fmt.Println("Connection failed:", err)
				}
				fmt.Println("Connected to:", newPeer)
				fmt.Println("> ")
			}
```

This far we get every moonshard devices discoverable and connected into one network

### Step 2 - setting up PubSub (publish/subscribe) and discover network topics

So, we have swarm of our peers connected to each other this far. 
However, we may want to _separate_ peers into groups, based on their _topics_. 
If we will __not__ do such thing - we will face with problem of broadcast - each time when we send message to the network - we send it to __all__ peers in network, so it may be working __slowly__ when we have to much peers.

First thing to do is initialize PubSub object as 
```
pb, err := pubsub.NewFloodsubWithProtocols(context.Background(), host, []protocol.ID{protocol.ID(cfg.ProtocolID)}, pubsub.WithMessageSigning(true), pubsub.WithStrictSignatureVerification(true))
```
this line initialize pubsub, using floodsub protocol (alternatively we could use gossipsub instead of), our libp2p host configuration from previous step, enable message signing and signature verification.

Then, we could easily subscribe to any topic using `pb.Subcribe(topic)` as :
```
subscription, err := pb.Subscribe(cfg.RendezvousString)
serviceTopic = cfg.RendezvousString
```
service topic - is a main general topic, which group _every_ peer in a network, so we could use it for pushing some service and important information.

After we did this - we want to know about _other_ topics in our network, so we could subscribe to them as well. 
For doing so - we send some service message to service topic (which means that we are asking every peer in network about their topics) as:
```
go readSub(subscription, incomingMessages)
go getNetworkTopics()
```

After that we could easily subscribe, publish and create new topics using such functions as 
` newTopic(), writeTopic(), readSub() `



## Building
Require go version >=1.12 , so make sure your `go version` is okay.  
**WARNING!** Building happen only when this project locates outside of GOPATH environment.

```bash
$ git clone https://github.com/MoonSHRD/p2chat
$ cd p2chat
$ go mod tidy
$ make
```

If you have trouble with go mod, then you can try clean source building
```
$ go get -v -d ./... # not sure that it's neccessary
```
Builded binary will be in `./cmd/`
