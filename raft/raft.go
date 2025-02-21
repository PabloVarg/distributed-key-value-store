package raft

import (
	"context"
	"log/slog"
	"time"

	"go.etcd.io/raft/v3"
)

type RaftNode struct {
	ticker   time.Ticker
	logger   *slog.Logger
	RaftNode raft.Node
	storage  *raft.MemoryStorage
}

func NewRaftNode(l *slog.Logger) RaftNode {
	return RaftNode{
		logger:  l,
		ticker:  *time.NewTicker(100 * time.Millisecond),
		storage: raft.NewMemoryStorage(),
	}
}

func (n *RaftNode) StartNode(ID uint64, peers []uint64) {
	c := &raft.Config{
		ID:              ID,
		ElectionTick:    10,
		HeartbeatTick:   1,
		Storage:         n.storage,
		MaxSizePerMsg:   4096,
		MaxInflightMsgs: 256,
	}

	p := []raft.Peer{{ID: ID}}
	for _, peer := range peers {
		p = append(p, raft.Peer{ID: peer})
	}
	n.logger.Info("raft: StartNode", "peers", p)

	n.RaftNode = raft.StartNode(c, p)
}

func (n RaftNode) Loop(ctx context.Context) {
	for {
		select {
		case <-n.ticker.C:
			n.RaftNode.Tick()
		case rd := <-n.RaftNode.Ready():
			n.storage.Append(rd.Entries)
			n.RaftNode.Advance()
		case <-ctx.Done():
			return
		}
	}
}
