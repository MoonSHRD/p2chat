package p2mobile

type Config struct {
	RendezvousString string // Unique string to identify group of nodes. Share this with your friends to let them connect with you
	ProtocolID       string // Sets a protocol id for stream headers
	ListenHost       string // The bootstrap node host listen address
	ListenPort       int    // Node listen port
}

func GetConfig() *Config {
	c := &Config{}

	c.RendezvousString = "meetme"
	c.ProtocolID = "/chat/1.1.0"
	c.ListenHost = "0.0.0.0"
	c.ListenPort = 4001

	return c
}
