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

type AppConf struct {
	Addr  string
	ID    uint64
	Peers []uint64
}

func main() {
	run(os.Stdout)
}

func run(w io.Writer) {
	l := NewLogger(w)
	n := raft.NewRaftNode(l)

	c := ReadConf()
	l.Info("read configuration", "conf", c)

	n.StartNode(c.ID, c.Peers)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()

		n.Loop(ctx)
		l.Info("shutting down", "ID", c.ID)
	}()

	srv := api.NewHTTPServer(l, c.Addr, n.RaftNode)
	wg.Add(1)
	go func() {
		defer wg.Done()

		l.Info("server listening on address", "addr", c.Addr)
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

			l.Info("server shutting down")
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

func ReadConf() AppConf {
	c := AppConf{
		Peers: make([]uint64, 0),
	}

	envID, ok := os.LookupEnv("ID")
	if !ok {
		panic("env ID is not configured")
	}

	ID, err := strconv.ParseUint(envID, 10, 64)
	if err != nil {
		panic("env ID is not a uint64")
	}
	c.ID = ID

	ReadPeersConf(&c)
	ReadAddr(&c)

	return c
}

func ReadPeersConf(c *AppConf) {
	envPeers, ok := os.LookupEnv("PEERS")
	if !ok {
		return
	}

	for peer := range strings.SplitSeq(envPeers, ",") {
		peerID, err := strconv.ParseUint(strings.TrimSpace(peer), 10, 64)
		if err != nil {
			panic("a peer in PEERS is not a uint64")
		}

		c.Peers = append(c.Peers, peerID)
	}
}

func ReadAddr(c *AppConf) {
	addr, ok := os.LookupEnv("API_ADDRESS")
	if !ok {
		c.Addr = ":8000"
		return
	}

	c.Addr = addr
}
