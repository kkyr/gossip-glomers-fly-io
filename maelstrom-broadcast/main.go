package main

import (
	"context"
	"encoding/json"
	"log"
	"sync"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

var (
	node *maelstrom.Node

	mu   sync.RWMutex
	seen map[float64]bool

	neighbourQueues map[string]*queue
)

func main() {
	node = maelstrom.NewNode()
	seen = make(map[float64]bool)
	neighbourQueues = make(map[string]*queue)

	node.Handle("broadcast", broadcast)
	node.Handle("read", read)
	node.Handle("topology", topology)

	log.Fatal(node.Run())
}

func broadcast(msg maelstrom.Message) error {
	var body map[string]any
	if err := json.Unmarshal(msg.Body, &body); err != nil {
		return err
	}

	response := map[string]string{
		"type": "broadcast_ok",
	}

	num := body["message"].(float64)
	mu.Lock()
	if seen[num] {
		mu.Unlock()
		return node.Reply(msg, response)
	}

	seen[num] = true
	mu.Unlock()

	for neighbour, q := range neighbourQueues {
		// do not broadcast to node that sent it to us
		if neighbour == msg.Src {
			continue
		}
		q.enqueue(num)
	}

	return node.Reply(msg, response)
}

func read(msg maelstrom.Message) error {
	var messages []float64
	mu.RLock()
	for k := range seen {
		messages = append(messages, k)
	}
	mu.RUnlock()

	return node.Reply(msg, map[string]any{
		"messages": messages,
		"type":     "read_ok",
	})
}

func topology(msg maelstrom.Message) error {
	var body struct {
		Topology map[string][]string `json:"topology"`
	}
	if err := json.Unmarshal(msg.Body, &body); err != nil {
		return err
	}

	neighbours := body.Topology[node.ID()]
	log.Printf("received topology, my neighbours are %+v", neighbours)

	for _, id := range neighbours {
		if _, ok := neighbourQueues[id]; !ok {
			q := newQueue(id)
			go q.start(context.Background())
			neighbourQueues[id] = q
		}
	}

	return node.Reply(msg, map[string]any{
		"type": "topology_ok",
	})
}
