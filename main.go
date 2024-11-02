package main

import (
	"log"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

func main() {
	log.Println("starting node")

	n := maelstrom.NewNode()
	n.Handle("broadcast", func(msg maelstrom.Message) error {
		log.Println("handling broadcast message")
		return nil
	})

	n.Handle("read", func(msg maelstrom.Message) error {
		log.Println("handling read message")
		return nil
	})

	n.Handle("topology", func(msg maelstrom.Message) error {
		log.Println("handling topology message")
		return nil
	})
}
