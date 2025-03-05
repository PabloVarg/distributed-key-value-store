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
	peers         []string
}

func NewRaftNode(l *slog.Logger, keyValueStore store.Store) RaftNode {
	return RaftNode{
		logger:        l,
		ticker:        *time.NewTicker(100 * time.Millisecond),
		storage:       raft.NewMemoryStorage(),
		keyValueStore: keyValueStore,
	}
}

func (n *RaftNode) StartNode(ID uint64, peers []string) {
	c := &raft.Config{
		ID:              ID,
		ElectionTick:    10,
		HeartbeatTick:   1,
		Storage:         n.storage,
		MaxSizePerMsg:   4096,
		MaxInflightMsgs: 256,
	}

	p := []raft.Peer{{ID: ID}}
	for i := range peers {
		p = append(p, raft.Peer{ID: uint64(i) + ID})
	}
	n.peers = peers
	n.logger.Info("raft: StartNode", "peers", peers)

	n.RaftNode = raft.StartNode(c, p)
}

func (n RaftNode) Loop(ctx context.Context) {
	for {
		select {
		case <-n.ticker.C:
			n.RaftNode.Tick()
		case rd := <-n.RaftNode.Ready():
			n.logger.Debug("node ready", "message", rd)

			n.saveState(rd)
			n.handleCommittedEntries(rd)

			n.RaftNode.Advance()
		case <-ctx.Done():
			return
		}
	}
}

func (n RaftNode) saveState(rd raft.Ready) {
	if !raft.IsEmptyHardState(rd.HardState) {
		n.storage.SetHardState(rd.HardState)
	}

	if !raft.IsEmptySnap(rd.Snapshot) {
		n.storage.ApplySnapshot(rd.Snapshot)
	}

	n.storage.Append(rd.Entries)
}

func (n RaftNode) handleCommittedEntries(rd raft.Ready) {
	if rd.CommittedEntries == nil {
		return
	}

	for entry := range slices.Values(rd.CommittedEntries) {
		switch entry.Type {
		case raftpb.EntryConfChange:
			n.logger.Debug("raft configuration change", "entry", entry)
			var cc raftpb.ConfChange
			cc.Unmarshal(entry.Data)
			n.RaftNode.ApplyConfChange(cc)
		case raftpb.EntryNormal:
			if entry.Data == nil {
				break
			}

			action, err := DecodeAction(n.logger, entry.Data)
			if err != nil {
				n.logger.Error("committed unreadable entry, ignoring", "data", entry.Data)
				continue
			}

			n.applyAction(action)
			n.logger.Info("applying committed entry", "action", action)
		}
	}
}

func (n RaftNode) applyAction(action StoreAction) {
	switch action.Action {
	case Put:
		n.keyValueStore.Put(action.Key, action.Value)
	case Delete:
		n.keyValueStore.Delete(action.Key)
	}
}
