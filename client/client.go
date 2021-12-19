/*
 * @file: client.go
 * @author: Jorge Quitério
 * @copyright (c) 2021 Jorge Quitério
 * @license: MIT
 */

package client

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
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

// ToByte converts the message to a byte array
func (msg *ClientMsg) ToByte() ([]byte, error) {
	b, err := json.Marshal(msg)
	if err != nil {
		return []byte{}, err
	}
	return b, nil
}
func (msg *ClientMsg) FromByte(b []byte) error {
	err := json.Unmarshal(b, msg)
	if err != nil {
		return err
	}
	return nil
}

func (client *HubClient) SendMsg(msg ClientMsg) error {
	b, err := msg.ToByte()
	if err != nil {
		return err
	}
	_, err = client.Conn.Write(b)
	return err
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
	rootcert, err := ioutil.ReadFile("ca.pem")
	if err != nil {
		panic(err)
	}
	roots := x509.NewCertPool()
	ok := roots.AppendCertsFromPEM(rootcert)
	if !ok {
		panic("failed to parse root certificate")
	}
	config := &tls.Config{
		Certificates:       []tls.Certificate{cert},
		RootCAs:            roots,
		InsecureSkipVerify: true,
	}
	tlsConn, err := tls.Dial("tcp", addr, config)
	if err != nil {
		panic(err)
	}
	client.Conn = tlsConn
}

func (client HubClient) PrintConnectionStatus() {
	log.Print(">>>>>>>>>>>>>>>> State <<<<<<<<<<<<<<<<")
	conn := client.Conn
	state := conn.ConnectionState()
	log.Println("Remote Address: ", conn.RemoteAddr())
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
