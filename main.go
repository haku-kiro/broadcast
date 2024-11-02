package main

import (
	"encoding/json"
	"log"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

func main() {
	log.Println("starting node")

	var messages = []any{}

	n := maelstrom.NewNode()
	n.Handle("broadcast", func(msg maelstrom.Message) error {
		log.Println("handling broadcast message")

		// deserialize the message, handle data message retrieve
		var body map[string]any
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		message := body["message"]
		messages = append(messages, message)

		response := map[string]any{
			"type": "broadcast_ok",
		}

		return n.Reply(msg, response)
	})

	n.Handle("read", func(msg maelstrom.Message) error {
		log.Println("handling read message")

		// Do you have to empty the messages after a read? Doesn't explicitly say so?
		response := map[string]any{
			"type":     "read_ok",
			"messages": messages,
		}

		return n.Reply(msg, response)
	})

	n.Handle("topology", func(msg maelstrom.Message) error {
		log.Println("handling topology message")

		response := map[string]any{
			"type": "topology_ok",
		}

		return n.Reply(msg, response)
	})

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}
