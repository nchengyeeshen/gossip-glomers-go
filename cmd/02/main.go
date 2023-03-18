package main

import (
	"log"

	"github.com/google/uuid"
	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

func main() {
	n := maelstrom.NewNode()

	n.Handle("generate", func(msg maelstrom.Message) error {
		return n.Reply(msg, map[string]any{
			"type": "generate_ok",
			// UUIDs have low chance of collisions and don't rely on any state
			// for generating new IDs, making it ideal for this use-case.
			"id":   uuid.NewString(),
		})
	})

	if err := n.Run(); err != nil {
		log.Fatalln("run:", err)
	}
}
