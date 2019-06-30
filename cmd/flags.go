package main

import (
	"flag"
)

type config struct {
	RendezvousString string
	ProtocolID       string
	listenHost       string
	listenPort       int
}

func parseFlags() *config {
	c := &config{}

	flag.StringVar(&c.RendezvousString, "rendezvous", "moonshard", "Unique string to identify group of nodes. Share this with your friends to let them connect with you")
	flag.StringVar(&c.listenHost, "wrapped_host", "0.0.0.0", "The bootstrap node wrapped_host listen address\n")
	flag.StringVar(&c.ProtocolID, "pid", "/chat/1.1.1", "Sets a protocol id for stream headers")
	flag.IntVar(&c.listenPort, "port", 4001, "node listen port")

	flag.Parse()
	return c
}
