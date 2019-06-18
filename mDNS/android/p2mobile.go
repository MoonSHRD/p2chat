package p2mobile

import (
	"bufio"
	"context"
	"crypto/rand"
	"fmt"
	"os"
	"strings"
	//	"github.com/MoonSHRD/p2chat/mDNS/android/flags"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-crypto"
	inet "github.com/libp2p/go-libp2p-net"
	protocol "github.com/libp2p/go-libp2p-protocol"
	"github.com/multiformats/go-multiaddr"
)

//
//
//   structs example in golang
//
// 	 type MobileLibp2p struct {
//  node *host.Host
//  }
//
//	func StartLibp2p() *MobileLibp2p {
//	  host, err := libp2p.New(ctx)
//	  return &MobileLibp2p{
//	    node: &host,
//	  }
//	}
//

//
type StreamApi struct {
	Potok inet.Stream
}

// TODO: global variable to get access to the stream outside of this daemon. In future should be replaced by mapping
// NOTE  if we want to access from Java to exactly this variable we need to make type of it as  "var P *StreamApi. " , where "*" means pointer.
// also if we want to export any function in Java - we should define its type by "*". Pure types can't be exported as is.
// also, if we actually will make type of P as a pointer, to get this var in android var - it possible will break the rest of the program due
// "invalid memory address or nil pointer dereference" so instead of this we will use special setters and getters for this variable, instead of direct access.
var P StreamApi

// NOTE:  pointer to the stream, but it is not a stream interface itself.
// If we want access stream class on Java side we should use exportable structure above
var Ptk *inet.Stream

func SetStreamApi(stream inet.Stream) {
	P.Potok = stream
}

// returning struct itself
func GetStreamApi() *StreamApi {
	//	return P.Potok
	return &P
}

// returning interface from a struct
func GetStreamApiInterface(ApiStruct *StreamApi) inet.Stream {
	streamInterface := ApiStruct.Potok
	return streamInterface
}

// TODO: useless code (duplicate) need to remove everything with Ptk, made this to test pointers bug
// NOTE:  here is work with global variables. Still don't sure about Java, so making two methods.
func GetStreamPointer() *inet.Stream {
	return Ptk
}
func SetStreamPointer(stream inet.Stream) {
	Ptk = &stream
}

// NOTE:
//    handleStream function is invoked in VHODYASHIE calls
//    at this moment of time we should RETURN some kind of a STREAM ID - which is
//    apparently is stream inet.Stream variable and put it into some kind of global variable (or a map for multiple connetctions in the future)
//    (as first I think to return a buffer, but in fact getting stream ID is a better idea)
//
//		After we done with it we could have a setted (and returned) global variable with a stream id (and exported getter for it)
//
//    Then we get user input from interface and invoke ... new WriteStream exported function which should have get stream ID and user input string as arguments
//		Then it funcion should itself make a new ReadWriter based on Stream ID and write string to stream, using rw.WriteString

func handleStream(stream inet.Stream) {
	fmt.Println("Got a new stream!")

	// Create a buffer stream for non blocking read and write.
	// NOTE: uncomment it for debug mode
	//	rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))

	// HACK: if we will use &stream pointer instead of stream interface we could get Ptk with pointer to a dinamic variable. Which means if stream interface will change - we will auto
	// switch to this new interface.
	// note - it can be multiple interfaces in one device, so, we MUST store some kind of stream ID in sturcture delayed in global mapping
	Ptk = &stream
	fmt.Println("stream pointer:")
	fmt.Println(Ptk)

	SetStreamApi(stream)

	// Check
	// TODO: remove this check in production build
	fmt.Println("stream interface:")
	fmt.Println(stream)
	fmt.Println("Checking setting stream")
	stream_struct := GetStreamApi()
	fmt.Println("Returning setted streamApi struct", stream_struct)
	stream_interface := GetStreamApiInterface(stream_struct)
	fmt.Println("Returning setted streamApi interface", stream_interface)

	// NOTE:
	// it is a good solution for desktop/console mode, when we whait for user input, but in android (where is no stdIn or direct console imput) we should avoid such invokation
	//	go readData(rw)
	//	go writeData(rw)

	// NOTE: this function should be invoked for debug/testing mode.
	// in normal mode this functions should be invoked only from StreamWriter
	//	StreamWriter(stream, "streamWriter Check")

	// 'stream' will stay open until you close it (or the other side closes it).
}

// this function should be invoked from java side to write messages in one perticular stream
func StreamWriter(potok *StreamApi, str string) {
	stream := potok.Potok
	if stream != nil {
		rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))
		msg := &str
		message := string(*msg)
		writeHandler(rw, message)
	}
}

func StreamReader(potok *StreamApi) string {
	stream := potok.Potok
	if stream != nil {
		rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))
		message := readHandler(rw)
		//	msg := string(*message)
		return message
	}
	return ""
}

func readData(rw *bufio.ReadWriter) {
	// NOTE: endless cycle here
	for {
		str, err := rw.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading from buffer")
			panic(err)
		}

		if str == "" {
			return
		}
		if str != "\n" {
			// Green console colour: 	\x1b[32m
			// Reset console colour: 	\x1b[0m
			fmt.Printf("\x1b[32m%s\x1b[0m> ", str)
		}

	}
}

// this function should take string as argument and write it to the buffer
// NOTE - this function is duplicate of writeData function. difference is in method of input (android doesn't have stdIn)
func writeHandler(rw *bufio.ReadWriter, str string) {
	// NOTE: os.StdIn is for console input
	//	strReader := bufio.NewReader(os.Stdin)
	msg := &str
	message := string(*msg)
	strReader := bufio.NewReader(strings.NewReader(message))

	// NOTE: endless cycle here
	for {

		sendData, err := strReader.ReadString('\n')

		// BUG: : somehow err here is always != nil (even if everything is ok) still don't know what to do here
		/*
			if err != nil {
				fmt.Println("Error reading from str")
				panic(err)
			}
		*/

		_, err = rw.WriteString(fmt.Sprintf("%s\n", sendData))
		if err != nil {
			fmt.Println("Error writing to buffer")
			panic(err)
		}
		err = rw.Flush()
		if err != nil {
			fmt.Println("Error flushing buffer")
			panic(err)
		}
	}
}

// this function should take string as argument and write it to the buffer
func readHandler(rw *bufio.ReadWriter) string {

	// NOTE: endless cycle here
	for {
		str, err := rw.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading from buffer")
			str = ""
		}

		if str == "" {
			return ""
		}
		if str != "\n" {
			// Green console colour: 	\x1b[32m
			// Reset console colour: 	\x1b[0m
			fmt.Printf("\x1b[32m%s\x1b[0m> ", str)
			//		msg := string(*str)
			return str
		}

	}
}

// NOTE: if this function will invoke from android side - app will crash.
func writeData(rw *bufio.ReadWriter) {
	stdReader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("> ")
		sendData, err := stdReader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading from stdin")
			panic(err)
		}

		_, err = rw.WriteString(fmt.Sprintf("%s\n", sendData))
		if err != nil {
			fmt.Println("Error writing to buffer")
			panic(err)
		}
		err = rw.Flush()
		if err != nil {
			fmt.Println("Error flushing buffer")
			panic(err)
		}
	}
}

func Start() {
	cfg := GetConfig()

	fmt.Printf("[*] Listening on: %s with port: %d\n", cfg.ListenHost, cfg.ListenPort)

	ctx := context.Background()
	r := rand.Reader

	// Creates a new RSA key pair for this host.
	prvKey, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, r)
	if err != nil {
		panic(err)
	}

	// 0.0.0.0 will listen on any interface device.
	sourceMultiAddr, _ := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/%s/tcp/%d", cfg.ListenHost, cfg.ListenPort))

	// libp2p.New constructs a new libp2p Host.
	// Other options can be added here.
	host, err := libp2p.New(
		ctx,
		libp2p.ListenAddrs(sourceMultiAddr),
		libp2p.Identity(prvKey),
	)

	if err != nil {
		panic(err)
	}

	// Set a function as stream handler.
	// This function is called when a peer initiates a connection and starts a stream with this peer. (Handle INcoming connections)
	host.SetStreamHandler(protocol.ID(cfg.ProtocolID), handleStream)

	fmt.Printf("\n[*] Your Multiaddress Is: /ip4/%s/tcp/%v/p2p/%s\n", cfg.ListenHost, cfg.ListenPort, host.ID().Pretty())

	peerChan := initMDNS(ctx, host, cfg.RendezvousString)

	peer := <-peerChan // will block untill we discover a peer
	fmt.Println("Found peer:", peer, ", connecting")

	if err := host.Connect(ctx, peer); err != nil {
		fmt.Println("Connection failed:", err)
	}

	// open a stream, this stream will be handled by handleStream other end		(Handle OUTcoming connections)
	stream, err := host.NewStream(ctx, peer.ID, protocol.ID(cfg.ProtocolID))

	if err != nil {
		fmt.Println("Stream open failed", err)
	} else {
		//	rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))

		SetStreamApi(stream)

		// TODO: remove for production build
		//	go writeData(rw)
		//	go readData(rw)

		fmt.Println("Connected to:", peer)
	}

	select {} //wait here
}
