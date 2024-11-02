package main

import (
	"encoding/json"
	"log"
	"sync"

	mapset "github.com/deckarep/golang-set/v2"
	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

type Container struct {
	mu   sync.Mutex
	data mapset.Set[int]
}

func (c *Container) addData(data int) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data.Add(data)
}

func (c *Container) addDataSlice(data []int) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data.Append(data...)
}

func main() {
	log.Println("starting node")

	var container = Container{}

	n := maelstrom.NewNode()
	n.Handle("broadcast", func(msg maelstrom.Message) error {
		log.Println("handling broadcast message")

		// deserialize the message, handle data message retrieve
		var body map[string]any
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		message := body["message"].(int)
		container.addData(message)

		data := container.data.ToSlice()
		sharedResponse := map[string]any{
			"type": "shared_data",
			"data": data,
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
		log.Println("received data to propagate")

		var body map[string]any
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		data := body["data"].([]int)
		container.addDataSlice(data)

		return nil
	})

	n.Handle("read", func(msg maelstrom.Message) error {
		log.Println("handling read message")

		// Do you have to empty the messages after a read? Doesn't explicitly say so?
		response := map[string]any{
			"type":     "read_ok",
			"messages": container.data,
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

	// Handle writing the data into the messages array,

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}
