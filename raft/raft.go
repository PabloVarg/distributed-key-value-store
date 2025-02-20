package raft

import (
	"context"
	"log/slog"

	"go.etcd.io/raft/v3"
)

type RaftNode struct {
	logger   *slog.Logger
	raftNode raft.Node
}

func NewRaftNode(l *slog.Logger) RaftNode {
	return RaftNode{
		logger: l,
	}
}

func (n *RaftNode) StartNode(ID uint64, peers []uint64) {
	c := &raft.Config{
		ID:              ID,
		ElectionTick:    10,
		HeartbeatTick:   1,
		Storage:         raft.NewMemoryStorage(),
		MaxSizePerMsg:   4096,
		MaxInflightMsgs: 256,
	}

	p := make([]raft.Peer, 0, 1)
	p = append(p, raft.Peer{ID: ID})
	for _, peer := range peers {
		p = append(p, raft.Peer{ID: peer})
	}

	n.raftNode = raft.StartNode(c, p)
}

func (n RaftNode) Loop(ctx context.Context) {
	for {
		select {
		case rd := <-n.raftNode.Ready():
			n.logger.Info("node ready", "rd", rd)
			n.raftNode.Advance()
		case <-ctx.Done():
			return
		}
	}
}
