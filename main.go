package main

import (
	"context"
	"io"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/pablovarg/distributed-key-value-store/raft"
)

func main() {
	l := NewLogger(os.Stdout)
	n := raft.NewRaftNode(l)

	n.StartNode()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		n.Loop(ctx)
	}()

	wg.Wait()
}

func NewLogger(w io.Writer) *slog.Logger {
	return slog.New(slog.NewJSONHandler(w, nil))
}
