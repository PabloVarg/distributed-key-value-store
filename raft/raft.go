package raft

import (
	"context"
	"log/slog"
	"slices"
	"time"

	"github.com/pablovarg/distributed-key-value-store/store"
	"go.etcd.io/raft/v3"
)

type RaftNode struct {
	ticker        time.Ticker
	logger        *slog.Logger
	RaftNode      raft.Node
	storage       *raft.MemoryStorage
	keyValueStore store.Store
}

func NewRaftNode(l *slog.Logger, keyValueStore store.Store) RaftNode {
	return RaftNode{
		logger:        l,
		ticker:        *time.NewTicker(100 * time.Millisecond),
		storage:       raft.NewMemoryStorage(),
		keyValueStore: keyValueStore,
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
			n.logger.Info("message", "message", rd.Entries)

			if rd.CommittedEntries != nil {
				for entry := range slices.Values(rd.CommittedEntries) {
					action, err := DecodeAction(n.logger, entry.Data)
					if err != nil {
						n.logger.Error("committed unreadable entry, ignoring", "data", entry.Data)
						continue
					}

					n.logger.Info("applying committed entry", "action", action)
					switch action.Action {
					case Put:
						n.keyValueStore.Put(action.Key, action.Value)
					case Delete:
						// TODO: Implement delete action
					}
				}
			}

			n.storage.Append(rd.Entries)
			n.RaftNode.Advance()
		case <-ctx.Done():
			return
		}
	}
}
