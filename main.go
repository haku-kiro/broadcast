package main

import (
	"context"
	"encoding/json"
	"log"
	"sync"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

type message struct {
	destination string
	body        map[string]any
}

type retry struct {
	cancel context.CancelFunc
	ch     chan message
}

func (r *retry) action(msg message) {
	r.ch <- msg
}

func (r *retry) close() {
	r.cancel()
}

func newRetry(n *maelstrom.Node, tries int) *retry {
	ch := make(chan message)
	ctx, cancel := context.WithCancel(context.Background())

	for i := 0; i < tries; i++ {
		go func() {
			for {
				select {
				case msg := <-ch:
					for {
						if err := n.Send(msg.destination, msg.body); err != nil {
							continue
						}
						break
					}
				case <-ctx.Done():
					return
				}
			}
		}()
	}

	return &retry{
		ch:     ch,
		cancel: cancel,
	}
}

type server struct {
	nodeId string
	node   *maelstrom.Node
	r      *retry

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
	var body map[string]any
	if err := json.Unmarshal(msg.Body, &body); err != nil {
		return err
	}

	id := int(body["message"].(float64))
	s.mu.Lock()
	if _, exists := s.data[id]; exists {
		s.mu.Unlock()
		return nil
	}
	s.data[id] = struct{}{}
	s.mu.Unlock()

	for _, node := range s.node.NodeIDs() {
		if node == s.nodeId {
			continue
		}

		s.r.action(message{
			destination: node,
			body:        body,
		})
	}

	return s.node.Reply(msg, map[string]any{
		"type": "broadcast_ok",
	})
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

func NewServer(n *maelstrom.Node, r *retry) *server {
	return &server{
		nodeId: n.ID(),
		node:   n,
		mu:     sync.RWMutex{},
		r:      r,
		data:   make(map[int]struct{}),
	}
}

func main() {
	n := maelstrom.NewNode()
	r := newRetry(n, 10)
	defer r.close()

	server := NewServer(n, r)

	n.Handle("broadcast", server.broadcastHandler)
	n.Handle("read", server.readHandler)
	n.Handle("topology", server.topologyHandler)

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}
