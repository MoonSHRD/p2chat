# p2chat

Examples of local chats on libp2p stack

Both examples work with automatic peer discovery
I've separated mdns from rendezvous point, so you can try both of methods as you  wish



## What's the main difference and how it could be implemented as MoonShard solution

Simply words, both methods use DHT for peerdiscovery

mDNS using micro-dns service, which means that routing node in the network should support
this service. Most of modern routers should support it, but not everyone
(i.g. mDNS will definetly not working at Moscow subways as we learned in fields testing)
Also not sure mDNS will work from mobile ad-hoc points, but have not tested it this far

Rendezvous point is better solution, and also should decrease battery consumption
Also it should better connect local chats with cloud (remote) nodes.
Rendezvous also great when thereare a lot of offline nodes behind NAT and it hard to connect with them

Probably will switch to SONM solution, but as far we are fully accomplish with libp2p stack


## How to build
require go version >=1.12 , so make sure your `go version` is okay

If it start yelling about go modules, try
```
export GO111MODULE=on
```
I've include it into Makefile, but not sure it will work correctly


### How to build rendezvous
From main repo run
```
> make deps
> cd ./rendezvous
> go build -o chat

```
### How to build rendezvous to windows devices
GOOS=windows GOARCH=amd64 go build -o chat_windows

### How to use rendezvous
Use two different terminal windows to run
```
./chat -listen /ip4/127.0.0.1/tcp/6666
./chat -listen /ip4/127.0.0.1/tcp/6668

```
Remember about NAT penetration!

### How to build mDNS
```
go get -v -d ./...
go build
```
### How to use mDNS  

Use two different terminal windows to run
```
./mDNS -port 6666
./mDNS -port 6667
```

## So how does it work?

1. **Configure a p2p host**
```go
ctx := context.Background()

// libp2p.New constructs a new libp2p Host.
// Other options can be added here.
host, err := libp2p.New(ctx)
```
[libp2p.New](https://godoc.org/github.com/libp2p/go-libp2p#New) is the constructor for libp2p node. It creates a host with given configuration.

2. **Set a default handler function for incoming connections.**

This function is called on the local peer when a remote peer initiate a connection and starts a stream with the local peer.
```go
// Set a function as stream handler.
host.SetStreamHandler("/chat/1.1.0", handleStream)
```

```handleStream``` is executed for each new stream incoming to the local peer. ```stream``` is used to exchange data between local and remote peer. This example uses non blocking functions for reading and writing from this stream.

```go
func handleStream(stream net.Stream) {

    // Create a buffer stream for non blocking read and write.
    rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))

    go readData(rw)
    go writeData(rw)

    // 'stream' will stay open until you close it (or the other side closes it).
}
```

3. **Find peers nearby using mdns**

Start [mdns discovery](https://godoc.org/github.com/libp2p/go-libp2p/p2p/discovery#NewMdnsService) service in host.

```go
peerChan := initMDNS(ctx, host, cfg.RendezvousString)
```
register [Notifee interface](https://godoc.org/github.com/libp2p/go-libp2p/p2p/discovery#Notifee) with service so that we get notified about peer discovery

```go
	n := &discoveryNotifee{}
	ser.RegisterNotifee(n)
```


4. **Open streams to peers found.**

Finally we open stream to the peers we found, as we find them

```go
	peer := <-peerChan // will block untill we discover a peer
	fmt.Println("Found peer:", peer, ", connecting")

	if err := host.Connect(ctx, peer); err != nil {
		fmt.Println("Connection failed:", err)
	}

	// open a stream, this stream will be handled by handleStream other end
	stream, err := host.NewStream(ctx, peer.ID, protocol.ID(cfg.ProtocolID))

	if err != nil {
		fmt.Println("Stream open failed", err)
	} else {
		rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))

		go writeData(rw)
		go readData(rw)
		fmt.Println("Connected to:", peer)
	}
```

### How to build Android module

```
cd ./mDNS/android/
gomobile bind -target=android -v
```

this will generate you `*.aar` and `*.jar`packages for android

then, open your project in android studio, go `File -> ProjectStructure -> modules -> new module -> Import aar/jar`
and then open "*.aar" file.

then you should press 'apply' and also add it as a dependancy to the project. You swicth for dependancy tab, then choose app module itself, then, in right window click add and add p2mobile module as a dependancy.

By default, you will able to invoke any experted functions (those one, which start with **C**apital letter in go code.


## What types and functions will be accesable from p2chat in my android app?

If you want be able to invoke any go functions from java side, you need to export them via renaming exported functions with Capital Letter like this `Start()`. Note, if you want to export functions with an unusual type, than you need to create a structure in go with this type and export it as well.

From java side just type `import p2mobile.*;` and then invoke like `P2mobile.Start()`
