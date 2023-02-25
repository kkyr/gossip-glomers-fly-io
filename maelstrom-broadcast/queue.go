package main

import (
	"context"
	"log"
	"time"
)

func newQueue(nodeID string) *queue {
	return &queue{
		nodeID:        nodeID,
		pending:       make(chan float64, 100),
		replyTimeout:  500 * time.Millisecond,
		batchDuration: 150 * time.Millisecond,
	}
}

// queue represents a message queue for a specific node. Nums can
// be enqueued onto the queue where they will eventually be sent to
// the node.
type queue struct {
	nodeID        string
	pending       chan float64
	replyTimeout  time.Duration
	batchDuration time.Duration
}

func (q *queue) enqueue(nums ...float64) {
	for _, num := range nums {
		q.pending <- num
	}
}

// start makes the queue listen for enqueue operations and sends
// them to the node. This is a blocking call and should typically
// be called asynchronously on a separate goroutine.
func (q *queue) start(ctx context.Context) {
	ticker := time.NewTicker(q.batchDuration)

	for {
		select {
		case <-ctx.Done():
			ticker.Stop()
			return
		case <-ticker.C:
			N := len(q.pending)
			if N == 0 { // no messages to send
				break
			}

			nums := make([]float64, N)
			for i := 0; i < N; i++ {
				nums[i] = <-q.pending
			}

			body := map[string]any{
				"type":     "broadcast",
				"messages": nums,
			}

			go func() {
				ctx, cancel := context.WithTimeout(ctx, q.replyTimeout)
				_, err := node.SyncRPC(ctx, q.nodeID, body)
				if err != nil {
					log.Printf("SyncRPC failed due to %v", err)
					// message failed or timed out -> requeue the messages so that we retry
					q.enqueue(nums...)
				}
				cancel()
			}()
		}
	}
}
