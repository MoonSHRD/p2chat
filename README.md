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

### How to build mDNS module for android and import it to Android Studio

go to the `./android` dir, then open terminal and type

` gomobile bind -target=android -v -d `
this wil generate `*aar` and `*jar` files, which you can use in your android studio project

When you want to import it to android IDE you should go to `File -> ProjectStructure -> Modules -> "+" add new module -> import *aar/*jar` and import your module. 

Then you go to `File -> ProjectStructure -> Dependencies` , select `app` and then in the right window add `p2mobile` as a dependency for your app. After it's done, gradle will auto generate everything you need.

## What types and functions will be accesable from p2chat in my android app?

If you want be able to invoke any go functions from java side, you need to export them via renaming exported functions with Capital Letter like this `Start()`. Note, if you want to export functions with an unusual type, than you need to create a structure in go with this type and export it as well.

From java side just type `import p2mobile.*;` and then invoke like `P2mobile.Start()`

### How to use mDNS  

Use two different terminal windows to run
```
./mDNS -port 6666
./mDNS -port 6667
```
