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
	"github.com/pablovarg/distributed-key-value-store/store"
	"go.etcd.io/raft/v3/raftpb"
)

type AppConf struct {
	Debug      bool
	Addr       string
	PeerAddr   string
	ID         uint64
	Peers      []string
	StorageDir string
}

func main() {
	run(os.Stdout)
}

func run(w io.Writer) {
	messagesTx := make(chan raftpb.Message)
	messagesRx := make(chan raftpb.Message)

	c := ReadConf()
	l := NewLogger(w, c.Debug)
	s := store.NewKeyValueStore()
	t := raft.NewTransport(
		l,
		c.PeerAddr,
		func(u uint64) string { return c.Peers[u-1] },
		messagesRx,
		messagesTx,
	)
	n := raft.NewRaftNode(l, s, messagesTx, t)

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

	wg.Add(1)
	go func() {
		defer wg.Done()

		n.StepToMessages(ctx)
	}()

	srv := api.NewHTTPServer(l, c.Addr, n.RaftNode, s)
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

	wg.Add(1)
	go func() {
		defer wg.Done()

		t.ListenAndServe(ctx)
	}()

	wg.Wait()
}

func NewLogger(w io.Writer, debug bool) *slog.Logger {
	level := slog.LevelInfo
	if debug {
		level = slog.LevelDebug
	}

	l := slog.New(slog.NewJSONHandler(w, &slog.HandlerOptions{
		Level: level,
	}))

	if debug {
		l.Info("DEBUG flag detected")
	}

	return l
}

func ReadConf() AppConf {
	c := AppConf{
		Peers: make([]string, 0),
	}

	envID := os.Getenv("ID")
	if strings.TrimSpace(envID) == "" {
		envID = "1"
	}

	ID, err := strconv.ParseUint(envID, 10, 64)
	if err != nil {
		panic("env ID is not a uint64")
	}
	c.ID = ID

	ReadPeersConf(&c)
	ReadAddr(&c)
	ReadPeerAddr(&c)
	ReadDebugFlag(&c)

	return c
}

func ReadDebugFlag(c *AppConf) {
	debug := os.Getenv("DEBUG")

	if strings.TrimSpace(strings.ToLower(debug)) != "true" {
		return
	}

	c.Debug = true
}

func ReadPeersConf(c *AppConf) {
	envPeers, ok := os.LookupEnv("PEERS")
	if !ok {
		return
	}

	c.Peers = strings.Split(envPeers, ",")
}

func ReadAddr(c *AppConf) {
	addr, ok := os.LookupEnv("API_ADDRESS")
	if !ok {
		c.Addr = ":8000"
		return
	}

	c.Addr = addr
}

func ReadPeerAddr(c *AppConf) {
	addr, ok := os.LookupEnv("PEER_ADDRESS")
	if !ok {
		c.PeerAddr = ":8001"
		return
	}

	c.PeerAddr = addr
}
