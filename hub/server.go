/*
 * @file: server.go
 * @author: Jorge Quitério
 * @copyright (c) 2021 Jorge Quitério
 * @license: MIT
 */

package hub

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
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

	cert, err := tls.LoadX509KeyPair("certs/server.pem", "certs/server.key")
	if err != nil {
		panic("hub.Serve.X509KeyPair: " + err.Error())
	}
	rootcert, err := ioutil.ReadFile("certs/ca.pem")
	if err != nil {
		panic(err)
	}
	roots := x509.NewCertPool()
	ok := roots.AppendCertsFromPEM(rootcert)
	if !ok {
		panic("failed to parse root certificate")
	}

	tlsconfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    roots,
	}

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
		go h.handleConn(conn.(*tls.Conn))
	}
}

func (h *Hub) handleConn(conn net.Conn) {
	defer conn.Close()
	//r := bufio.NewReader(conn)
	for {

		printConnState(conn.(*tls.Conn))

		nconn := conn.(*tls.Conn)
		buf := make([]byte, 1024)
		n, err := nconn.Read(buf)
		if err != nil {
			log.Println(err)
			break
		}
		fmt.Println("Received: " + string(buf[:n]))
		received := buf[:n]

		// sending back
		nconn.Write([]byte("OK"))

		var clientMsg ClientMsg

		// receive client message
		err = json.Unmarshal(received, &clientMsg)
		if err != nil {
			log.Println(err)
			return
		}
		client, ok := h.GetClient(clientMsg.ClientID)
		if !ok {
			if clientMsg.MsgType == "subscribe" {
				log.Println("New client: " + clientMsg.ClientID)
				cli := NewClient(clientMsg.ClientID, []string{clientMsg.Topic})
				if ok := h.AddClient(cli); ok {
					log.Println("Client " + clientMsg.ClientID + " subscribed to " + clientMsg.Topic)
				}
			}
			b, _ := json.Marshal(h.Clients)
			fmt.Println("Sending back: " + string(b))
			io.WriteString(nconn, string(b))
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
		nconn.Write([]byte(b))
		fmt.Println(string(b))
		fmt.Println("Client " + clientMsg.ClientID + " not found")
	}
}

func printConnState(conn *tls.Conn) {
	log.Print(">>>>>>>>>>>>>>>> State <<<<<<<<<<<<<<<<")
	state := conn.ConnectionState()
	log.Printf("Version: %x", state.Version)
	log.Printf("HandshakeComplete: %t", state.HandshakeComplete)
	log.Printf("DidResume: %t", state.DidResume)
	log.Printf("CipherSuite: %x", state.CipherSuite)
	log.Printf("NegotiatedProtocol: %s", state.NegotiatedProtocol)
	log.Printf("NegotiatedProtocolIsMutual: %t", state.NegotiatedProtocolIsMutual)

	log.Print("Certificate chain:")
	for i, cert := range state.PeerCertificates {
		subject := cert.Subject
		issuer := cert.Issuer
		log.Printf(" %d s:/C=%v/ST=%v/L=%v/O=%v/OU=%v/CN=%s", i, subject.Country, subject.Province, subject.Locality, subject.Organization, subject.OrganizationalUnit, subject.CommonName)
		log.Printf("   i:/C=%v/ST=%v/L=%v/O=%v/OU=%v/CN=%s", issuer.Country, issuer.Province, issuer.Locality, issuer.Organization, issuer.OrganizationalUnit, issuer.CommonName)
	}
	log.Print(">>>>>>>>>>>>>>>> State End <<<<<<<<<<<<<<<<")
}
