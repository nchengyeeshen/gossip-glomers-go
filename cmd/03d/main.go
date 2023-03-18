package main

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
	"golang.org/x/exp/maps"
)

func main() {
	srv := NewServer(maelstrom.NewNode())

	if err := srv.Run(); err != nil {
		log.Fatalln("run:", err)
	}
}

func NewServer(node *maelstrom.Node) *Server {
	s := &Server{
		node:     node,
		topo:     make(map[string][]string),
		messages: make(map[int]struct{}),
	}

	node.Handle("broadcast", s.BroadcastHandler)
	node.Handle("read", s.ReadHandler)
	node.Handle("topology", s.TopologyHandler)

	return s
}

type Server struct {
	node *maelstrom.Node

	topo   map[string][]string
	topoMu sync.RWMutex

	messagesMu sync.RWMutex
	messages   map[int]struct{}
}

func (s *Server) Run() error {
	return s.node.Run()
}

func (s *Server) BroadcastHandler(msg maelstrom.Message) error {
	var body struct {
		Message int `json:"message"`
		ID      int `json:"msg_id"`
	}
	if err := json.Unmarshal(msg.Body, &body); err != nil {
		return err
	}

	// Check for the presence of a message ID. Node.Send expects no reply, so
	// message ID is not included.
	if body.ID != 0 {
		s.node.Reply(msg, map[string]any{
			"type": "broadcast_ok",
		})
	}

	if s.messageExists(body.Message) {
		return nil
	}

	go func(src string, msg int) {
		s.storeMessage(body.Message)
		s.broadcastToNeighbours(src, body.Message)
	}(msg.Src, body.Message)

	return nil
}

func (s *Server) ReadHandler(msg maelstrom.Message) error {
	return s.node.Reply(msg, map[string]any{
		"type":     "read_ok",
		"messages": s.readMessages(),
	})
}

func (s *Server) TopologyHandler(msg maelstrom.Message) error {
	var body struct {
		Topology map[string][]string `json:"topology"`
	}
	if err := json.Unmarshal(msg.Body, &body); err != nil {
		return err
	}

	s.setTopo(body.Topology)

	return s.node.Reply(msg, map[string]any{
		"type": "topology_ok",
	})
}

func (s *Server) broadcastToNeighbours(src string, msg int) {
	for _, neighbour := range s.neighbours(s.node.ID()) {
		// Don't broadcast to the node that sent us this message. It already has
		// the message!
		if neighbour == src {
			continue
		}

		go func(dest string) {
			var (
				acked bool
				mu    sync.Mutex
			)

			// Try for up to 100 times.
			for i := 1; !acked && i <= 100; i++ {
				s.node.RPC(
					dest,
					map[string]any{
						"type":    "broadcast",
						"message": msg,
					},
					func(msg maelstrom.Message) error {
						mu.Lock()
						defer mu.Unlock()
						acked = true
						return nil
					},
				)

				// Linear backoff.
				time.Sleep(time.Duration(i) * time.Second)
			}

		}(neighbour)
	}
}

func (s *Server) readMessages() []int {
	s.messagesMu.RLock()
	defer s.messagesMu.RUnlock()
	return maps.Keys(s.messages)
}

func (s *Server) storeMessage(msg int) {
	s.messagesMu.Lock()
	defer s.messagesMu.Unlock()
	s.messages[msg] = struct{}{}
}

func (s *Server) messageExists(msg int) bool {
	s.messagesMu.RLock()
	defer s.messagesMu.RUnlock()
	_, ok := s.messages[msg]
	return ok
}

func (s *Server) setTopo(m map[string][]string) {
	s.topoMu.Lock()
	defer s.topoMu.Unlock()
	s.topo = m
}

func (s *Server) neighbours(node string) []string {
	s.topoMu.RLock()
	defer s.topoMu.RUnlock()
	return s.topo[node]
}
