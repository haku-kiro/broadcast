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

	mu   sync.RWMutex
	data map[int]struct{}
}

func (s *server) getIds() []int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	keys := make([]int, 0, len(s.data))
	for k := range s.data {
		keys = append(keys, k)
	}

	return keys
}

func (s *server) broadcastHandler(msg maelstrom.Message) error {
	response := map[string]any{
		"type": "broadcast_ok",
	}

	var body map[string]any
	if err := json.Unmarshal(msg.Body, &body); err != nil {
		return err
	}

	message := int(body["message"].(float64))
	s.mu.Lock()
	if _, exists := s.data[message]; exists {
		s.mu.Unlock()
		return nil
	}
	s.data[message] = struct{}{}
	s.mu.Unlock()

	for _, node := range s.node.NodeIDs() {
		if node == s.nodeId {
			continue
		}

		s.node.Send(node, body)
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
		mu:     sync.RWMutex{},
		data:   make(map[int]struct{}),
	}
}

func main() {
	n := maelstrom.NewNode()
	server := NewServer(n)

	n.Handle("broadcast", server.broadcastHandler)
	n.Handle("read", server.readHandler)
	n.Handle("topology", server.topologyHandler)

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}
