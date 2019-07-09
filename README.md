# P2Chat
P2Chat - is a core local messenger library, which based on Libp2p stack.

P2Chat basicaly supports discovery through **mDNS** service and support messaging via **PubSub**

It supports next features:
- devices autodiscovery by `Rendezvous string`
- topic list exchanging between peers
- autoconnect group chats by `PubSub`
- default signing and validating messages (crypto)
- crossplatform


## Building
Require go version >=1.12 , so make sure your `go version` is okay

```bash
$ git clone https://github.com/MoonSHRD/p2chat
$ cd p2chat
$ go mod tidy
$ go get -v -d ./... # not sure that it's neccessary
$ make
```
Builded binary will be in `./cmd/`
