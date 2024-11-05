package main

import (
	"encoding/json"
	"log"
	"sync"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

type server struct {
	nodeId string
	node   *maelstrom.Node

	mu   sync.Mutex
	data map[int]struct{}
}

func (s *server) getIds() []int {
	s.mu.Lock()
	defer s.mu.Unlock()

	keys := make([]int, 0, len(s.data))
	for k := range s.data {
		keys = append(keys, k)
	}

	return keys
}

func (s *server) addId(d int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data[d] = struct{}{}
}

func (s *server) broadcastHandler(msg maelstrom.Message) error {
	var body map[string]any
	if err := json.Unmarshal(msg.Body, &body); err != nil {
		return err
	}

	message := int(body["message"].(float64))
	s.addId(message)

	sharedResponse := map[string]any{
		"type":    "shared_data",
		"message": message,
	}

	for _, node := range s.node.NodeIDs() {
		if node == s.nodeId {
			continue
		}

		s.node.RPC(node, sharedResponse, func(msg maelstrom.Message) error {
			return nil
		})
	}

	response := map[string]any{
		"type": "broadcast_ok",
	}

	return s.node.Reply(msg, response)
}

func (s *server) readHandler(msg maelstrom.Message) error {
	response := map[string]any{
		"type":     "read_ok",
		"messages": s.getIds(),
	}

	return s.node.Reply(msg, response)
}

func (s *server) topologyHandler(msg maelstrom.Message) error {
	return s.node.Reply(msg, map[string]any{
		"type": "topology_ok",
	})
}

func NewServer(n *maelstrom.Node) *server {
	return &server{
		nodeId: n.ID(),
		node:   n,
		mu:     sync.Mutex{},
		data:   make(map[int]struct{}),
	}
}

func main() {
	n := maelstrom.NewNode()
	server := NewServer(n)

	n.Handle("broadcast", server.broadcastHandler)
	n.Handle("read", server.readHandler)
	n.Handle("topology", server.topologyHandler)

	n.Handle("shared_data", func(msg maelstrom.Message) error {
		var body map[string]any
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		data := int(body["message"].(float64))
		server.addId(data)

		body["type"] = "shared_data_ok"

		return n.Reply(msg, body)
	})

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}
