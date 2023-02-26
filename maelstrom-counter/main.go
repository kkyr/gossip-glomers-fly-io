package main

import (
	"context"
	"encoding/json"
	"log"
	"math/rand"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

const counterKey = "counter"

var (
	node *maelstrom.Node
	kv   *maelstrom.KV
)

func main() {
	node = maelstrom.NewNode()
	kv = maelstrom.NewSeqKV(node)

	node.Handle("add", add)
	node.Handle("read", read)

	log.Fatal(node.Run())
}

func add(msg maelstrom.Message) error {
	body := make(map[string]any)
	if err := json.Unmarshal(msg.Body, &body); err != nil {
		return err
	}

	delta := int(body["delta"].(float64))

	for {
		err := addDelta(delta)
		if err != nil && maelstrom.ErrorCode(err) != maelstrom.PreconditionFailed {
			return err
		} else if err == nil {
			break
		}
	}

	return node.Reply(msg, map[string]any{
		"type": "add_ok",
	})
}

func read(msg maelstrom.Message) error {
	var v value
	var err error

	for {
		v, err = getValue()
		if err != nil {
			return err
		}

		v2 := newValue(v.Counter)
		err = kv.CompareAndSwap(context.Background(), counterKey, v, v2, true)
		if err != nil && maelstrom.ErrorCode(err) != maelstrom.PreconditionFailed {
			return err
		} else if err == nil {
			break
		}
	}

	return node.Reply(msg, map[string]any{
		"type":  "read_ok",
		"value": v.Counter,
	})
}

func addDelta(delta int) error {
	v, err := getValue()
	if err != nil {
		return err
	}

	v2 := newValue(v.Counter)
	err = kv.CompareAndSwap(context.Background(), counterKey, v, v2, true)
	if err != nil {
		return err
	}

	v3 := newValue(v2.Counter + delta)
	err = kv.CompareAndSwap(context.Background(), counterKey, v2, v3, true)
	if err != nil {
		return err
	}

	return nil
}

func getValue() (value, error) {
	v, err := kv.Read(context.Background(), counterKey)
	if err != nil {
		if maelstrom.ErrorCode(err) == maelstrom.KeyDoesNotExist {
			return newValue(0), nil
		}
		return value{}, err
	}

	m := v.(map[string]interface{})
	return value{
		Counter: int(m["counter"].(float64)),
		Rnd:     int(m["rnd"].(float64)),
	}, nil
}

type value struct {
	Counter int `json:"counter"`
	Rnd     int `json:"rnd"`
}

// returns a new value with a random Rnd value.
func newValue(counter int) value {
	return value{
		Counter: counter,
		Rnd:     randInt(),
	}
}

func randInt() int {
	return int(rand.Int63())
}
