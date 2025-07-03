package chatroom

import (
	"context"
	"encoding/json"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/peer"
)

const (
	NamePrefix        = "chatroom:"
	MessageBufferSize = 100 // Size of the buffer for chat messages
)

type Message struct {
	Message    string
	SenderID   string
	SenderNick string
}

type Room struct {
	Messages chan *Message
	ctx      context.Context
	ps       *pubsub.PubSub
	topic    *pubsub.Topic
	sub      *pubsub.Subscription

	roomName string
	self     peer.ID
	nick     string
}

func Join(ctx context.Context, ps *pubsub.PubSub, roomName string, selfID peer.ID, nick string) (*Room, error) {
	topic, err := ps.Join(topicName(roomName))
	if err != nil {
		return nil, err
	}

	sub, err := topic.Subscribe()
	if err != nil {
		return nil, err
	}

	cr := &Room{
		ctx:      ctx,
		ps:       ps,
		topic:    topic,
		sub:      sub,
		self:     selfID,
		nick:     nick,
		roomName: roomName,
		Messages: make(chan *Message, MessageBufferSize),
	}

	go cr.readLoop()
	return cr, nil
}

func (cr *Room) Publish(msg string) error {
	m := &Message{
		Message:    msg,
		SenderID:   cr.self.String(),
		SenderNick: cr.nick,
	}

	msgBytes, err := json.Marshal(m)
	if err != nil {
		return err
	}

	if err := cr.topic.Publish(cr.ctx, msgBytes); err != nil {
		return err
	}
	return nil
}

func (cr *Room) RoomName() string {
	return cr.roomName
}

func (cr *Room) SelfNick() string {
	return cr.nick
}

func (cr *Room) ListPeers() []peer.ID {
	return cr.ps.ListPeers(topicName(cr.roomName))
}

func (cr *Room) Context() context.Context {
	return cr.ctx
}

func (cr *Room) readLoop() {
	for {
		msg, err := cr.sub.Next(cr.ctx)
		if err != nil {
			close(cr.Messages)
			return
		}

		if msg.ReceivedFrom == cr.self {
			// Ignore messages sent by ourselves
			continue
		}

		cm := new(Message)

		if err := json.Unmarshal(msg.Data, cm); err != nil {
			continue
		}

		cr.Messages <- cm
	}
}

func topicName(roomName string) string {
	return NamePrefix + roomName
}
