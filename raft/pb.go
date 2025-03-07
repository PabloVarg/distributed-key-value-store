package raft

import (
	"context"
	"errors"
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
		defer t.logger.Debug("transport", "step", "exit send routine")

		t.logger.Debug("transport", "step", "start send routine")
		t.Send()
	}()

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

func (t Transport) Send() {
	for message := range t.messagesRxChan {
		t.logger.Debug("transport", "step", "send message", "message", message)
		conn, err := net.Dial("tcp", t.peers(message.To))
		if err != nil {
			continue
		}

		msg, err := proto.Marshal(&message)
		if err != nil {
			continue
		}

		t.logger.Info("transport", "step", "send message", "message", msg)
		if _, err := conn.Write(msg); err != nil {
			continue
		}

		conn.Close()
	}
}

func (t Transport) Listen() {
	l, err := net.Listen("tcp", t.addr)
	if err != nil {
		panic(err) // TODO: Better error handling
	}
	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			panic(err) // TODO: Better error handling
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
			case errors.Is(err, net.ErrClosed):
				return
			default:
				t.logger.Error("transport", "step", "reading", "err", err)
			}
		}
		b = b[:n]
		t.logger.Info("transport", "step", "receive bytes", "bytes", b)

		var msg raftpb.Message
		if err := proto.Unmarshal(b, &msg); err != nil {
			t.logger.Error("transport", "step", "unmarshaling", "err", err)
		}

		t.logger.Info("transport", "step", "receive message", "message", msg)
		t.messagesTxChan <- msg
	}
}
