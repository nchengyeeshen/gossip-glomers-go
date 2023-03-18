package main

import (
	"encoding/json"
	"log"
	"sync"

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
	topoMu sync.Mutex

	messagesMu sync.RWMutex
	messages   map[int]struct{}
}

func (s *Server) Run() error {
	return s.node.Run()
}

func (s *Server) BroadcastHandler(msg maelstrom.Message) error {
	var body struct {
		Message int `json:"message"`
	}
	if err := json.Unmarshal(msg.Body, &body); err != nil {
		return err
	}

	s.storeMessage(body.Message)

	return s.node.Reply(msg, map[string]any{
		"type": "broadcast_ok",
	})
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

func (s *Server) setTopo(m map[string][]string) {
	s.topoMu.Lock()
	defer s.topoMu.Unlock()
	s.topo = m
}
