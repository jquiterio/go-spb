/*
 * @file: main.go
 * @author: Jorge Quitério
 * @copyright (c) 2021 Jorge Quitério
 * @license: MIT
 */

package hub

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/go-redis/redis/v8"
	"github.com/jquiterio/go-spb/utils"
)

type HubServer struct {
	Addr        string
	Port        string
	Certificate utils.AuthCert
}

type Hub struct {
	Registry     *redis.Client
	Subscription *redis.PubSub
	Server       *HubServer
	Clients      map[string]HubClient
}

type HubClient struct {
	ID     string
	Topics []string
}

type ClientMsg struct {
	ClientID string
	MsgID    string
	MsgType  string
	Topic    string
	Data     interface{}
}

func NewClient(id string, topics []string) HubClient {
	return HubClient{
		ID:     id,
		Topics: topics,
	}
}

func (h *Hub) GetClient(id string) (HubClient, bool) {
	client := h.Clients[id]
	if client.ID != "" {
		return client, true
	}
	return HubClient{}, false
}

func (h *Hub) AddClient(hc HubClient) bool {
	if hc.ID != "" {
		h.Clients[hc.ID] = hc
		return true
	}
	return false
}

type Message *redis.Message

var ctx = context.Background()

func NewHub() Hub {

	redis_addr := os.Getenv("REDIS_ADDR")
	if redis_addr == "" {
		fmt.Println("REDIS_ADDR not set. Using default: localhost:6379")
		redis_addr = "localhost:6379"
	}
	redis_pass := os.Getenv("REDIS_PASS")
	if redis_pass == "" {
		fmt.Println("REDIS_PASS not set. Not using password")
		redis_pass = ""
	}
	redis_db, _ := strconv.ParseInt(os.Getenv("REDIS_DB"), 10, 32)
	if redis_db == 0 {
		fmt.Println("REDIS_DB not set or set to 0. Using default: 0")
		redis_db = 0
	}
	redis := redis.NewClient(&redis.Options{
		Addr:     redis_addr,
		Password: redis_pass,
		DB:       int(redis_db),
	})
	return Hub{
		Registry: redis,
		Server:   NewHubServer(),
	}
}

func NewHubServer() *HubServer {
	addr := os.Getenv("HUB_ADDR")
	if addr == "" {
		fmt.Println("HUB_ADDR not set. Using default: localhost")
		addr = "localhost"
	}
	port := os.Getenv("HUB_PORT")
	if port == "" {
		fmt.Println("HUB_PORT not set. Using default: 8083")
		port = "8083"
	}
	certs, err := utils.GenCert()
	if err != nil {
		panic(err)
	}
	return &HubServer{
		Addr:        addr,
		Port:        port,
		Certificate: certs,
	}
}

func (hub *Hub) Publish(topic string, msg interface{}) error {
	reg := hub.Registry
	if err := reg.Publish(ctx, topic, msg).Err(); err != nil {
		return err
	}
	return nil
}

func (hub *Hub) Subscribe(topics ...string) {
	reg := hub.Registry
	hub.Subscription = reg.Subscribe(ctx, topics...)
}

func (hub *Hub) Unsubscribe(topics ...string) error {
	sub := hub.Subscription
	return sub.Unsubscribe(ctx, topics...)
}

func (hub *Hub) Close() error {
	reg := hub.Registry
	err := reg.Close()
	if err != nil {
		return err
	}
	return nil
}

func (hub *Hub) GetMessage(topic string) (interface{}, error) {
	sub := hub.Subscription
	for {
		return sub.ReceiveMessage(ctx)
	}
}

func (hub *Hub) GetSubscribedMessages(topic string) *redis.Message {
	sub := hub.Subscription
	for {
		msg, _ := sub.ReceiveMessage(ctx)
		for _, c := range hub.Clients {
			for _, t := range c.Topics {
				if t == msg.Channel {
					return msg
				}
			}
		}
	}
}
