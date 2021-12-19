/*
 * @file: client.go
 * @author: Jorge Quitério
 * @copyright (c) 2021 Jorge Quitério
 * @license: MIT
 */

package client

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"

	"github.com/gofrs/uuid"
)

type HubClient struct {
	ID     string
	Topics []string
	Conn   *tls.Conn
}

type ClientMsg struct {
	ClientID string
	MsgID    string
	MsgType  string
	Topic    string
	Data     interface{}
}

// ToByte converts the message to a byte array
func (msg *ClientMsg) ToByte() ([]byte, error) {
	b, err := json.Marshal(msg)
	if err != nil {
		return []byte{}, err
	}
	return b, nil
}

func (client HubClient) NewSubscription(topic string) ClientMsg {
	client.Topics = append(client.Topics, topic)
	return ClientMsg{
		ClientID: uuid.Must(uuid.NewV4()).String(),
		MsgID:    uuid.Must(uuid.NewV4()).String(),
		MsgType:  "subscribe",
		Topic:    topic,
	}
}

func NewHubClient(id string, topics []string) *HubClient {
	if id == "" {
		id = uuid.Must(uuid.NewV4()).String()
	}
	return &HubClient{
		ID:     id,
		Topics: topics,
	}
}

func (c *HubClient) Disconnect() error {
	return nil
}

func (client *HubClient) NewTLSConnection(addr string, cert tls.Certificate) {
	// USER/PASS AUTH
	if addr == "" {
		addr = "localhost:8083"
	}
	// TLS CONN
	// cert, err := tls.LoadX509KeyPair(clientCert, clientKey)
	// if err != nil {
	// 	log.Fatalln(err)
	// }
	config := &tls.Config{
		Certificates:       []tls.Certificate{cert},
		InsecureSkipVerify: false,
	}
	tlsConn, err := tls.Dial("tcp", addr, config)
	if err != nil {
		panic(err)
	}
	client.Conn = tlsConn
}

func (c *HubClient) Subscribe(msg ClientMsg) error {
	b, err := msg.ToByte()
	if err != nil {
		return err
	}
	_, err = c.Conn.Write(b)
	return err
}

func (c *HubClient) ReceiveMsg() {
	for {
		buf := make([]byte, 1024)
		n, err := c.Conn.Read(buf)
		if err != nil {
			log.Println(err)
			return
		}
		fmt.Println(string(buf[:n]))
	}
}
