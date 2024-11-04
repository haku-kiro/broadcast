package main

import (
	"encoding/json"
	"errors"
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

func main() {
	var container = Container{}
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
		container.addData(int(message))

		// Create the sharedResponse to then send the data to other nodes
		sharedResponse := map[string]any{
			"type":    "shared_data",
			"message": message,
		}

		slog.Info("broadcast to all nodes")
		for _, node := range n.NodeIDs() {
			if node == n.ID() {
				continue
			}

			n.RPC(node, sharedResponse, func(msg maelstrom.Message) error {
				var response map[string]any
				if err := json.Unmarshal(msg.Body, &response); err != nil {
					return err
				}

				if msg.Type() != "shared_data_ok" {
					return errors.New("unexpected response")
				}
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

		message := body["message"].(float64)
		container.addData(int(message))

		body["type"] = "shared_data_ok"

		return n.Reply(msg, body)
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
		slog.Error(err.Error())
	}
}
