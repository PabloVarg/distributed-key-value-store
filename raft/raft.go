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

func (n *RaftNode) StartNode() {
	c := &raft.Config{
		ID:              0x01,
		ElectionTick:    10,
		HeartbeatTick:   1,
		Storage:         raft.NewMemoryStorage(),
		MaxSizePerMsg:   4096,
		MaxInflightMsgs: 256,
	}

	n.raftNode = raft.StartNode(c, []raft.Peer{
		{ID: 0x01},
	})
}

func (n RaftNode) Loop(ctx context.Context) {
	for {
		select {
		case rd := <-n.raftNode.Ready():
			n.logger.Info("node ready", "rd", rd)
		}
	}
}
