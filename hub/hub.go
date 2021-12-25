/*
 * @file: hub.go
 * @author: Jorge Quitério
 * @copyright (c) 2021 Jorge Quitério
 * @license: MIT
 */

package hub

import (
	"context"
	"encoding/json"

	"github.com/go-redis/redis/v8"
	"github.com/gofrs/uuid"
	"github.com/jquiterio/mhub/config"
)

var ctx = context.Background()

type Hub struct {
	Subscribers []Subscriber
	Topics      []string
	Registry    *redis.Client
}

type message struct {
	SubscriberID string `json:"subscriber_id"`
	MsgID        string `json:"msg_id"`
	MsgType      string `json:"msg_type"`
	Topic        string `json:"topic"`
	Msg          interface {
	} `json:"msg"`
}

type Subscriber struct {
	ID     string
	Topics []string
}

func (m *message) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"subscriber_id": m.SubscriberID,
		"msg_id":        m.MsgID,
		"topic":         m.Topic,
		"msg":           m.Msg,
	}
}

func NewHub() *Hub {
	conf := config.Config
	return &Hub{
		Subscribers: make([]Subscriber, 0),
		Topics:      make([]string, 0),
		Registry: redis.NewClient(&redis.Options{
			Addr: conf.Redis.Addr,
			DB:   conf.Redis.DB,
		}),
	}
}

func (h *Hub) Subscribe(sub Subscriber) {
	h.Subscribers = append(h.Subscribers, sub)
	h.addTopicFromSubscribers()
}

func (h *Hub) removeTopicFromSubscribers() {
	for _, sub := range h.Subscribers {
		for i, t := range sub.Topics {
			if !h.HasTopic(t) {
				sub.Topics = append(sub.Topics[:i], sub.Topics[i+1:]...)
			}
		}
	}
}

func (h *Hub) HasTopic(topic string) bool {
	for _, t := range h.Topics {
		if t == topic {
			return true
		}
	}
	return false
}

func (h *Hub) Unsubscribe(sub *Subscriber, topics []string) (ok bool) {
	for _, topic := range topics {
		sub.RemoveTopic(topic)
	}
	h.removeTopicFromSubscribers()
	return true
}

func (h *Hub) GetSubscriber(id string) *Subscriber {
	for _, sub := range h.Subscribers {
		if sub.ID == id {
			return &sub
		}
	}
	return nil
}

func (h *Hub) addTopicFromSubscribers() {
	for _, sub := range h.Subscribers {
		h.Topics = append(h.Topics, sub.Topics...)
	}
}

func (m *message) FromMap(msg map[string]interface{}) error {
	m.SubscriberID = msg["subscriber_id"].(string)
	m.MsgID = msg["msg_id"].(string)
	m.Msg = msg["msg"]
	m.Topic = msg["topic"].(string)
	return nil
}

func (m *message) ToJson() ([]byte, error) {
	return json.Marshal(m.ToMap())
}

func Newmessage(sub Subscriber, topic string, msg interface{}) *message {
	return &message{
		SubscriberID: sub.ID,
		MsgID:        uuid.Must(uuid.NewV4()).String(),
		Topic:        topic,
		Msg:          msg,
	}
}

func NewSubscriber(topics ...string) *Subscriber {
	return &Subscriber{
		ID:     uuid.Must(uuid.NewV4()).String(),
		Topics: topics,
	}
}

func (s Subscriber) HasTopic(topic string) bool {
	for _, t := range s.Topics {
		if t == topic {
			return true
		}
	}
	return false
}

func (s *Subscriber) AddTopic(topic string) {
	s.Topics = append(s.Topics, topic)
}

func (s *Subscriber) RemoveTopic(topic string) {
	for i, t := range s.Topics {
		if t == topic {
			s.Topics = append(s.Topics[:i], s.Topics[i+1:]...)
			return
		}
	}
}

func (h *Hub) Publish(msg message) error {
	return h.Registry.Publish(ctx, msg.Topic, msg.Msg).Err()
}
