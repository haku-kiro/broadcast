package main

import (
	"encoding/json"
	"log"
	"sync"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

type Container struct {
	mu   sync.Mutex
	data []any
}

func (c *Container) addData(data any) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data = append(c.data, data)
}

func main() {
	var container = Container{}

	n := maelstrom.NewNode()
	n.Handle("broadcast", func(msg maelstrom.Message) error {
		var body map[string]any
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		message := body["message"]
		container.addData(message)

		sharedResponse := map[string]any{
			"type":    "shared_data",
			"message": message,
		}

		for _, node := range n.NodeIDs() {
			if node == n.ID() {
				continue
			}
			n.Send(node, sharedResponse)
		}

		response := map[string]any{
			"type": "broadcast_ok",
		}

		return n.Reply(msg, response)
	})

	n.Handle("shared_data", func(msg maelstrom.Message) error {
		var body map[string]any
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		message := body["message"]
		container.addData(message)

		return nil
	})

	n.Handle("read", func(msg maelstrom.Message) error {
		response := map[string]any{
			"type":     "read_ok",
			"messages": container.data,
		}

		return n.Reply(msg, response)
	})

	n.Handle("topology", func(msg maelstrom.Message) error {
		response := map[string]any{
			"type": "topology_ok",
		}

		return n.Reply(msg, response)
	})

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}
