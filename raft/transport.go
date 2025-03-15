package raft

import (
	"context"

	"go.etcd.io/raft/v3/raftpb"
)

type Transporter interface {
	Send(message raftpb.Message, to string) raftpb.Message
	ListenAndServe(ctx context.Context)
}
