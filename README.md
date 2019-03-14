# p2chat

Examples of local chats on libp2p stack

Both examples work with automatic peer discovery
I've separated mdns from rendezvous point, so you can try both of methods as you  wish


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
