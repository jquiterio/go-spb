/*
 * @file: server.go
 * @author: Jorge Quitério
 * @copyright (c) 2021 Jorge Quitério
 * @license: MIT
 */

package hub

import (
	"bufio"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/jquiterio/go-spb/utils"
)

// func getSubscriptionRequest(conn *net.Conn) {
// 	buf := make([]byte, 1024)
// 	_, err := conn.Read(buf)
// 	if err != nil {
// 		log.Println(err)
// 	}
// 	if string(buf) == "subscribe:" {
// 		SendMsg(conn, []byte("OK"))
// 	} else {
// 		SendMsg(conn, []byte("NOK"))
// 	}
// }

func receiveMsg(conn net.Conn, buf []byte) {
	//buf := make([]byte, 1024)
	_, err := conn.Read(buf)
	if err != nil {
		log.Println(err)
	}
	fmt.Println(string(buf))
}

func SendMsg(conn *tls.Conn, msg []byte) {
	_, err := conn.Write(msg)
	if err != nil {
		log.Println(err)
	}
}

func (h *Hub) Serve() {

	var ip, port string
	hub_addr := os.Getenv("HUB_ADDR")
	hub_port := os.Getenv("HUB_PORT")

	cert, err := tls.LoadX509KeyPair("server.pem", "server.key")
	if err != nil {
		panic("hub.Serve.X509KeyPair: " + err.Error())
	}

	tlsconfig := &tls.Config{Certificates: []tls.Certificate{cert}}
	tlsconfig.InsecureSkipVerify = false

	if hub_addr == "" {
		ip = utils.GetIPAddr()
		if ip == "" {
			ip = "0.0.0.0"
		}
	} else {
		ip = hub_addr
	}
	if hub_port == "" {
		port = "8083"
	} else {
		port = hub_port
	}

	ls := ip + ":" + port
	ln, err := tls.Listen("tcp", ls, tlsconfig)
	if err != nil {
		panic(err)
	}

	defer ln.Close()
	fmt.Println("Listening on " + ls)
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		go h.handleConn(conn)
	}
}

func (h *Hub) handleConn(conn net.Conn) {
	defer conn.Close()
	//r := bufio.NewReader(conn)
	for {

		var clientMsg ClientMsg

		buf, err := bufio.NewReader(conn).ReadBytes('\n')
		if err != nil {
			log.Println(err)
			return
		}
		// receive client message
		err = json.Unmarshal(buf, &clientMsg)
		if err != nil {
			log.Println(err)
			return
		}
		fmt.Println(clientMsg)
		client, ok := h.GetClient(clientMsg.ClientID)
		if !ok {
			if clientMsg.MsgType == "subscribe" {
				cli := NewClient(clientMsg.ClientID, []string{clientMsg.Topic})
				if ok := h.AddClient(cli); ok {
					fmt.Println("Client " + clientMsg.ClientID + " subscribed to " + clientMsg.Topic)
				}
			}
		} else {
			if clientMsg.MsgType == "subscribe" {
				client.Topics = append(client.Topics, clientMsg.Topic)
				fmt.Println("Client " + clientMsg.ClientID + " subscribed to " + clientMsg.Topic)
				h.Subscribe(clientMsg.Topic)
			}
			if clientMsg.MsgType == "unsubscribe" {
				for i, topic := range client.Topics {
					if topic == clientMsg.Topic {
						client.Topics = append(client.Topics[:i], client.Topics[i+1:]...)
						fmt.Println("Client " + clientMsg.ClientID + " unsubscribed from " + clientMsg.Topic)
						break
					}
				}
			}
			if clientMsg.MsgType == "publish" {
				fmt.Println("Client " + clientMsg.ClientID + " published to " + clientMsg.Topic)
				h.Publish(clientMsg.Topic, clientMsg.Data)
			}
			for _, tpc := range client.Topics {
				retMsg := h.GetSubscribedMessages(tpc)
				conn.Write([]byte(retMsg.Payload))
			}
		}

		// print Client JSON
		b, err := json.Marshal(clientMsg)
		if err != nil {
			log.Println(err)
		}
		fmt.Println(string(b))
		fmt.Println("Client " + clientMsg.ClientID + " not found")
	}
}
