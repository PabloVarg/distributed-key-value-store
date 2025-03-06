package raft

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"sync"

	"github.com/gogo/protobuf/proto"
	"go.etcd.io/raft/v3/raftpb"
)

type Transport struct{
    logger *slog.Logger
    addr string
    peers []string
    messagesRxChan <-chan raftpb.Message
    messagesTxChan chan<- raftpb.Message
}

func NewTransport(l *slog.Logger, addr string, peers []string, messagesRx <-chan raftpb.Message, messagesTx chan<- raftpb.Message) Transport {
    return Transport{
        logger: l,
        addr: addr,
        peers: peers,
        messagesRxChan: messagesRx,
        messagesTxChan: messagesTx,
    }
}

func (t Transport) ListenAndServe(ctx context.Context) {
    var wg sync.WaitGroup

    wg.Add(1)
    go func () {
        defer wg.Done()

        t.Send()
    }()

    wg.Add(1)
    go func ()  {
        defer wg.Done()

        t.Listen()
    }()

    <-ctx.Done()
}

func (t Transport) Send() {
    for message := range t.messagesRxChan {
        t.logger.Debug("transport send", "message", message)
        conn, err := net.Dial("tcp", t.peers[message.To])
        if err != nil {
            panic(err) // TODO: Better error handling
        }

        msg, err := proto.Marshal(&message)
        if err != nil {
            panic(err) // TODO: Better error handling
        }

        t.logger.Info("transport rx", "msg", msg)
        if _, err := conn.Write(msg); err != nil {
            panic(err) // TODO: Better error handling
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
        _, err := conn.Read(b)
        if err != nil {
            switch {
            case errors.Is(err, net.ErrClosed):
                return
            default:
                panic(err) // TODO: Better error handling
            }
        }

        var msg raftpb.Message
        if err := proto.Unmarshal(b, &msg); err != nil {
            panic(err) // TODO: Better error handling
        }

        t.logger.Info("transport tx", "msg", msg)
        t.messagesTxChan <- msg
    }
}
