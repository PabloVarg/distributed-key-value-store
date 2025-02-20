package main

import (
	"context"
	"io"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/pablovarg/distributed-key-value-store/raft"
)

func main() {
	l := NewLogger(os.Stdout)
	n := raft.NewRaftNode(l)

	ID, peersIDs := ReadConf()
	l.Info("read configuration", "ID", ID, "peers", peersIDs)

	n.StartNode(ID, peersIDs)

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

func ReadConf() (ID uint64, peers []uint64) {
	envID, ok := os.LookupEnv("ID")
	if !ok {
		panic("env ID is not configured")
	}

	ID, err := strconv.ParseUint(envID, 10, 64)
	if err != nil {
		panic("env ID is not a uint64")
	}

	envPeers, ok := os.LookupEnv("PEERS")
	if !ok {
		return
	}

	for _, peer := range strings.Split(envPeers, ",") {
		peerID, err := strconv.ParseUint(strings.TrimSpace(peer), 10, 64)
		if err != nil {
			panic("a peer in PEERS is not a uint64")
		}

		peers = append(peers, peerID)
	}

	return
}
