# p2chat

## What is this and how do I do rest of my life about it?
p2hcat - is a core local messenger library, which based on Libp2p stack.

p2chat basicly supports discovery through **mdns** service and support messaging via **PubSub**

It supports next features:
- devices autodiscovery by `Rendez-vous string`
- topic list exchanging between peers
- autoconnect group chats by `PubSub`
- default signing and validating messages (crypto) *[validation is temporary off, due to the incorrect signing messages on Android]*
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
Builded binary will be in ./cmd/