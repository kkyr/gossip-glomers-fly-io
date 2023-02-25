package main

import (
	"context"
	"log"
	"time"
)

func newQueue(nodeID string) *queue {
	return &queue{
		nodeID:  nodeID,
		pending: make(chan float64),
		timeout: 500 * time.Millisecond,
	}
}

// queue represents a message queue for a specific node. Nums can
// be enqueued onto the queue where they will eventually be sent to
// the node.
type queue struct {
	nodeID  string
	pending chan float64
	timeout time.Duration
}

func (q *queue) enqueue(num float64) {
	q.pending <- num
}

// start makes the queue listen for enqueue operations and sends
// them to the node. This is a blocking call and should typically
// be called asynchronously on a separate goroutine.
func (q *queue) start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case num := <-q.pending:
			log.Printf("sending num %v to node %q", num, q.nodeID)

			body := map[string]any{
				"type":    "broadcast",
				"message": num,
			}

			go func() {
				ctx, cancel := context.WithTimeout(ctx, q.timeout)
				_, err := node.SyncRPC(ctx, q.nodeID, body)
				if err != nil {
					log.Printf("SyncRPC failed due to %v", err)
					// message failed or timed out -> requeue the message so that we retry
					q.enqueue(num)
				}
				cancel()
			}()
		}
	}
}
