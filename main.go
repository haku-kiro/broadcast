package main

import (
	"context"
	"encoding/json"
	"log"
	"log/slog"
	"os"
	"sync"
	"time"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

const maxRetries = 100

type message struct {
	destination string
	body        map[string]any
}

type server struct {
	nodeId  string
	node    *maelstrom.Node
	context context.Context

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

	slog.Info("sending to all nodes")
	for _, node := range s.node.NodeIDs() {
		if node == s.nodeId {
			continue
		}

		slog.Info("sending to node", "node", node, "from", s.node.ID())

		err := s.syncRPC(node, body)
		if err != nil {
			slog.Error("error when trying to broadcast to other nodes going into retry", "error", err)
			for i := 0; i < maxRetries; i++ {
				if err := s.syncRPC(node, body); err != nil {
					time.Sleep(time.Duration(i) * time.Second)
					continue
				}
				break
			}
			return err
		}
	}

	return s.node.Reply(msg, map[string]any{
		"type": "broadcast_ok",
	})
}

func (s *server) syncRPC(dst string, body map[string]any) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_, err := s.node.SyncRPC(ctx, dst, body)
	return err
}

func (s *server) readHandler(msg maelstrom.Message) error {
	response := map[string]any{
		"type":     "read_ok",
		"messages": s.getIds(),
	}

	return s.node.Reply(msg, response)
}

func (s *server) initHandler(_ maelstrom.Message) error {
	// ID only becomes valid after init
	s.nodeId = s.node.ID()
	return nil
}

func (s *server) topologyHandler(msg maelstrom.Message) error {
	return s.node.Reply(msg, map[string]any{
		"type": "topology_ok",
	})
}

func NewServer(n *maelstrom.Node) *server {
	return &server{
		nodeId:  n.ID(),
		node:    n,
		mu:      sync.RWMutex{},
		context: context.Background(),
		data:    make(map[int]struct{}),
	}
}

func main() {
	logFile, err := os.OpenFile("log.json", os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	if err != nil {
		panic(err)
	}

	logger := slog.New(slog.NewJSONHandler(logFile, nil))
	slog.SetDefault(logger)

	slog.Info("creating node")
	n := maelstrom.NewNode()

	slog.Info("creating server")
	server := NewServer(n)

	slog.Info("adding handlers")
	n.Handle("broadcast", server.broadcastHandler)
	n.Handle("read", server.readHandler)
	n.Handle("topology", server.topologyHandler)
	n.Handle("init", server.initHandler)

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}
