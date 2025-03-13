package raft

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net"
	"sync"

	"github.com/golang/protobuf/proto"
	"go.etcd.io/raft/v3/raftpb"
)

type PeersLookup func(uint64) string

type Transport struct {
	logger         *slog.Logger
	addr           string
	peers          PeersLookup
	messagesRxChan <-chan raftpb.Message
	messagesTxChan chan<- raftpb.Message
}

func NewTransport(
	l *slog.Logger,
	addr string,
	peers PeersLookup,
	messagesRx <-chan raftpb.Message,
	messagesTx chan<- raftpb.Message,
) Transport {
	return Transport{
		logger:         l,
		addr:           addr,
		peers:          peers,
		messagesRxChan: messagesRx,
		messagesTxChan: messagesTx,
	}
}

func (t Transport) ListenAndServe(ctx context.Context) {
	t.logger.Debug("transport", "step", "starting routines")
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer t.logger.Debug("transport", "step", "exit listen routine")

		t.logger.Debug("transport", "step", "start listen routine")
		t.Listen()
	}()

	wg.Wait()

	t.logger.Debug("transport", "step", "exit routines")
}

func (t Transport) Send(message raftpb.Message, to string) raftpb.Message {
	t.logger.Debug("transport", "step", "send message", "message", message)
	conn, err := net.Dial("tcp", to)
	if err != nil {
		t.logger.Debug("transport", "step", "send message", "err", err, "to", to)
		return raftpb.Message{}
	}
	defer conn.Close()

	msg, err := proto.Marshal(&message)
	if err != nil {
		t.logger.Debug("transport", "step", "send message", "err", err)
		return raftpb.Message{}
	}

	t.logger.Info("transport", "step", "send message", "message", msg)
	if _, err := conn.Write(msg); err != nil {
		t.logger.Debug("transport", "step", "send message", "err", err)
		return raftpb.Message{}
	}

	return message
}

func (t Transport) Listen() {
	l, err := net.Listen("tcp", t.addr)
	if err != nil {
		t.logger.Error("transport", "step", "listen", "err", err)
	}
	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			t.logger.Error("transport", "step", "listen", "err", err)
			continue
		}

		go t.ReadMessages(conn)
	}
}

func (t Transport) ReadMessages(conn net.Conn) {
	defer conn.Close()

	for {
		b := make([]byte, 1024)
		n, err := conn.Read(b)
		if err != nil {
			switch {
			case errors.Is(err, io.EOF):
				return
			default:
				t.logger.Error("transport", "step", "reading", "err", err)
				return
			}
		}
		b = b[:n]

		var msg raftpb.Message
		if err := proto.Unmarshal(b, &msg); err != nil {
			t.logger.Error("transport", "step", "unmarshaling", "err", err)
		}

		t.logger.Info("transport", "step", "receive message", "message", msg)
		t.messagesTxChan <- msg
	}
}
