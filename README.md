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
require go version >=1.11 , so make sure your `go version` is okay

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

### How to build Android module

``` 
cd ./mDNS/android/
gomobile bind -target=android -v
```

this will generate you `*.aar` and `*.jar`packages for android 

then, open your project in android studio, go File-> ProjectStructure -> modules -> new module -> Import aar/ jar
and then open `*aar` file.

then you should press 'apply' and also add it as a dependancy to the project. You swicth for dependancy tab, then choose app module itself, then, in right window click add and add p2mobile module as a dependancy.

By default, you will able to invoke any experted functions (those one, which start with **C**apital letter in go code.
### How to use mDNS  

Use two different terminal windows to run
```
./mDNS -port 6666
./mDNS -port 6667
```
