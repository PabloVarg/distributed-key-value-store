package raft

import (
	"context"
	"log/slog"
	"slices"
	"time"

	"github.com/pablovarg/distributed-key-value-store/store"
	"go.etcd.io/raft/v3"
	"go.etcd.io/raft/v3/raftpb"
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
			n.logger.Debug("node ready", "message", rd)
			if rd.CommittedEntries != nil {
				for entry := range slices.Values(rd.CommittedEntries) {
					if entry.Type == raftpb.EntryConfChange {
						var cc raftpb.ConfChange
						cc.Unmarshal(entry.Data)
						n.RaftNode.ApplyConfChange(cc)
						continue
					}

					if entry.Data == nil || entry.Type != raftpb.EntryNormal {
						continue
					}

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
						n.keyValueStore.Delete(action.Key)
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
