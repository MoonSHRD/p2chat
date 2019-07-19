# P2Chat
P2Chat - is a core local messenger library, which based on Libp2p stack.

P2Chat basicaly supports discovery through **mDNS** service and support messaging via **PubSub**

It supports following features:
- devices autodiscovery by `Rendezvous string`
- topic list exchanging between peers
- autoconnect group chats by `PubSub`
- default signing and validating messages (crypto)
- crossplatform


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
