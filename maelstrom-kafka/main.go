package main

import (
	"encoding/json"
	"log"
	"sync"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

var (
	node *maelstrom.Node

	dataMu sync.RWMutex
	data   map[string][]int

	committedOffsetsMu sync.RWMutex
	committedOffsets   map[string]int
)

func main() {
	node = maelstrom.NewNode()
	data = make(map[string][]int)
	committedOffsets = make(map[string]int)

	node.Handle("send", send)
	node.Handle("poll", poll)
	node.Handle("commit_offsets", commitOffsets)
	node.Handle("list_committed_offsets", listCommittedOffsets)

	log.Fatal(node.Run())
}

func send(msg maelstrom.Message) error {
	var body struct {
		Key string `json:"key"`
		Msg int    `json:"msg"`
	}
	if err := json.Unmarshal(msg.Body, &body); err != nil {
		return err
	}

	dataMu.Lock()
	if _, ok := data[body.Key]; !ok {
		data[body.Key] = make([]int, 0, 8)
	}

	data[body.Key] = append(data[body.Key], body.Msg)
	offset := len(data[body.Key]) - 1
	dataMu.Unlock()

	response := map[string]any{
		"type":   "send_ok",
		"offset": offset,
	}

	return node.Reply(msg, response)
}

func poll(msg maelstrom.Message) error {
	var body struct {
		Offsets map[string]int `json:"offsets"`
	}
	if err := json.Unmarshal(msg.Body, &body); err != nil {
		return err
	}

	msgs := make(map[string][][]int)

	dataMu.RLock()
	for key, offset := range body.Offsets {
		msgs[key] = make([][]int, 0)
		if _, ok := data[key]; ok && offset < len(data[key]) {
			pair := []int{offset, data[key][offset]}
			msgs[key] = append(msgs[key], pair)
		}
	}
	dataMu.RUnlock()

	response := map[string]any{
		"type": "poll_ok",
		"msgs": msgs,
	}

	return node.Reply(msg, response)
}

func commitOffsets(msg maelstrom.Message) error {
	var body struct {
		Offsets map[string]int `json:"offsets"`
	}
	if err := json.Unmarshal(msg.Body, &body); err != nil {
		return err
	}

	committedOffsetsMu.Lock()
	for key, offset := range body.Offsets {
		committedOffsets[key] = offset
	}
	committedOffsetsMu.Unlock()

	response := map[string]any{
		"type": "commit_offsets_ok",
	}

	return node.Reply(msg, response)
}

func listCommittedOffsets(msg maelstrom.Message) error {
	var body struct {
		Keys []string `json:"offsets"`
	}
	if err := json.Unmarshal(msg.Body, &body); err != nil {
		return err
	}

	offsets := make(map[string]int)

	committedOffsetsMu.RLock()
	for _, key := range body.Keys {
		offsets[key] = committedOffsets[key]
	}
	committedOffsetsMu.RUnlock()

	response := map[string]any{
		"type":    "list_committed_offsets_ok",
		"offsets": offsets,
	}

	return node.Reply(msg, response)
}
