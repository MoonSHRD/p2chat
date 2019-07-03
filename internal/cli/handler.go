package cli

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/MoonSHRD/p2chat/api"
	"github.com/MoonSHRD/p2chat/pkg/p2chat"
)

type Handler struct {
	node  *p2chat.Node
	topic string
}

func NewHandler(node *p2chat.Node, topic string) *Handler {
	h := &Handler{
		node:  node,
		topic: topic,
	}
	return h
}

func (h *Handler) TextMessage(msg p2chat.TextMessage) {
	fmt.Printf("%s \x1b[32m%s\x1b[0m> ", msg.From, msg.Body)
}

func (h *Handler) Peer(addr p2chat.PeerAddr) {
	fmt.Println("\nConnected to:", addr)
}

func (h *Handler) ReadStdin(ctx context.Context) {
	stdReader := bufio.NewReader(os.Stdin)
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		fmt.Print("> ")
		text, err := stdReader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				return
			}

			log.Printf("Error reading from stdin: %s", err)
			return
		}
		if text == "\n" {
			continue
		}
		message := &api.BaseMessage{
			Body: text,
			Flag: api.FLAG_GENERIC_MESSAGE,
		}

		err = h.node.Publish(h.topic, message)
		if err != nil {
			log.Printf("Error occurred when publishing: %s, err")
		}
	}
}
