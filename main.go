package main

import (
	"encoding/json"
	"log/slog"
	"os"
	"sync"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

type Container struct {
	mu   sync.Mutex
	data []int
}

func (c *Container) addData(data int) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data = append(c.data, data)
}

func (c *Container) getSlice() []int {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.data
}

func NewContainer() *Container {
	return &Container{
		mu:   sync.Mutex{},
		data: make([]int, 0),
	}
}

func main() {
	var container = NewContainer()
	logFile, err := os.OpenFile("log.json", os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	if err != nil {
		panic(err)
	}

	logger := slog.New(slog.NewJSONHandler(logFile, nil))
	slog.SetDefault(logger)

	n := maelstrom.NewNode()
	n.Handle("broadcast", func(msg maelstrom.Message) error {
		// unmarshal the body to get the message data
		var body map[string]any
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		message := body["message"].(float64)

		d := int(message)
		container.addData(d)

		// Create the sharedResponse to then send the data to other nodes
		// Sending all of the data to each node after getting a new element,
		// Should look to only do this when there is a network issue.
		sharedResponse := map[string]any{
			"type":    "shared_data",
			"message": d,
		}

		slog.Info("broadcast to all nodes except yourself")
		for _, node := range n.NodeIDs() {
			if node == n.ID() {
				continue
			}

			n.RPC(node, sharedResponse, func(msg maelstrom.Message) error {
				return nil
			})
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

		data := body["message"].(float64)
		container.addData(int(data))

		body["type"] = "shared_data_ok"

		return n.Reply(msg, body)
	})

	n.Handle("read", func(msg maelstrom.Message) error {
		response := map[string]any{
			"type":     "read_ok",
			"messages": container.getSlice(),
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
		slog.Error(err.Error())
	}
}
