package main

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/pablovarg/distributed-key-value-store/api"
	"github.com/pablovarg/distributed-key-value-store/raft"
)

func main() {
	run(os.Stdout)
}

func run(w io.Writer) {
	l := NewLogger(w)
	n := raft.NewRaftNode(l)

	ID, peersIDs := ReadConf()
	l.Info("read configuration", "ID", ID, "peers", peersIDs)

	n.StartNode(ID, peersIDs)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()

		n.Loop(ctx)
		l.Info("shutting down", "ID", ID)
	}()

	srv := api.NewHTTPServer(l, ":8000", n.RaftNode)
	wg.Add(1)
	go func() {
		defer wg.Done()

		l.Info("server listening on port", "port", 8000)
		if err := srv.ListenAndServe(); err != nil {
			switch {
			case errors.Is(err, http.ErrServerClosed):
				l.Info("http server closing")
			default:
				l.Error("error on server ListenAndServe", "err", err)
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		select {
		case <-ctx.Done():
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			if err := srv.Shutdown(ctx); err != nil {
				l.Error("error on server Shutdown", "err", err)
			}
		}
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
