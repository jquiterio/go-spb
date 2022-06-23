/*
 * @file: msg.go
 * @author: Jorge Quitério
 * @copyright (c) 2022 Jorge Quitério
 * @license: MIT
 */

package mhub

import (
	"github.com/jquiterio/uuid"
)

type Message struct {
	SubscriberID string `json:"subscriber_id"`
	ID           string `json:"id"`
	Topic        string `json:"topic"`
	Payload      string `json:"payload"`
}

func NewMessage(subscriberID, topic, payload string) *Message {
	return &Message{
		SubscriberID: subscriberID,
		ID:           uuid.NewV4().String(),
		Topic:        topic,
		Payload:      payload,
	}
}

func ParseMessage(subscriber, topic, payload string) (*Message, error) {
	m := NewMessage(subscriber, topic, payload)
	return m, nil
}

// func (m *Message) FromMap(msg map[string]interface{}) error {
// 	m.SubscriberID = msg["subscriber_id"].(string)
// 	m.ID = msg["id"].(string)
// 	m.Data = msg["msg"]
// 	m.Topic = msg["topic"].(string)
// 	return nil
// }

// func (m *Message) ToJSON() ([]byte, error) {
// 	return json.Marshal(m.ToMap())
// }
