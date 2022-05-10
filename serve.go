/*
 * @file: serve.go
 * @author: Jorge Quitério
 * @copyright (c) 2022 Jorge Quitério
 * @license: MIT
 */

package mhub

import (
	"bufio"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/valyala/fasthttp"
)

type (
	SubscriberRequestConfig struct {
		Skipper      Skipper
		TargetHeader string
	}
	Skipper func(ctx *fiber.Ctx) bool
)

var subscriberHeader string = "X-Subscriber-ID"

func HandlerSubscriberRequest() fiber.Handler {
	sConfig := SubscriberRequestConfig{
		Skipper:      func(ctx *fiber.Ctx) bool { return false },
		TargetHeader: subscriberHeader,
	}
	return func(c *fiber.Ctx) error {
		if sConfig.Skipper(c) {
			return c.Next()
		}
		res := c.Response()
		//req_headers := c.GetReqHeaders()
		//sid := req_headers[sConfig.TargetHeader]
		sid := c.Get(subscriberHeader)
		if sid == "" {
			return c.Status(401).JSON("Please provide a valid subscriber id")
		}
		res.Header.Set(subscriberHeader, sid)
		return c.Next()
	}
}

func (h *Hub) getSubscriberFromRequest(id string) *Subscriber {
	return h.GetSubscriber(id)
}

func genCertError(err error) {
	fmt.Errorf("Error reading ca.pem: %s", err)
	fmt.Errorf("Please Proceed to generate the certificate")
	fmt.Errorf("As follows:")
	fmt.Errorf("openssl genrsa -out ca.key 2048")
	fmt.Errorf("openssl req -new -x509 -key ca.key -out ca.pem -days 3650 -subj '/CN=ca'")
	fmt.Errorf("openssl genrsa -out server.key 2048")
	fmt.Errorf("openssl req -new -nodes -key server.key -out server.csr -subj '/CN=server'")
	fmt.Errorf("openssl x509 -req -in server.csr -CA ca.pem -CAkey ca.key -CAcreateserial -out server.pem -days 3650")
	fmt.Errorf("openssl genrsa -out client.key 2048")
	fmt.Errorf("openssl req -new -nodes -key client.key -out client.csr -subj '/CN=client'")
	fmt.Errorf("openssl x509 -req -in client.csr -CA ca.pem -CAkey ca.key -CAcreateserial -out client.pem -days 3650")
}

func (h *Hub) newServer() *fiber.App {
	app := fiber.New()

	app.Use(logger.New())
	app.Use(HandlerSubscriberRequest())
	app.Use(recover.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET, POST, PUT, DELETE, OPTIONS",
		AllowHeaders: "Origin, Accept, Content-Type, " + subscriberHeader,
	}))

	app.Get("/", h.getMessages)
	app.Post("/publish/:topic", h.publishToTopic)
	app.Post("/subscribe", h.subscribeToTopics)
	app.Post("/unsubscribe", h.unsubscribeTopics)
	app.Get("/:topic", h.getMessageTopic)

	return app
}

func (h *Hub) Serve() {
	conf := GetFromEnvOrDefault()

	app := h.newServer()
	hub_addr := conf.Hub.Addr + ":" + conf.Hub.Port

	if conf.Hub.Secure {
		caPem, err := ioutil.ReadFile("ca.pem")
		if err != nil {
			genCertError(err)
			return
		}
		rootca := x509.NewCertPool()
		if ok := rootca.AppendCertsFromPEM(caPem); !ok {
			panic("Failed to parse root certificate")
		}
		tlsconfig := &tls.Config{
			Certificates:             nil,
			MinVersion:               tls.VersionTLS12,
			PreferServerCipherSuites: true,
			CipherSuites: []uint16{
				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
				tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_RSA_WITH_AES_256_CBC_SHA,
			},
			ClientAuth: tls.RequireAndVerifyClientCert,
			ClientCAs:  rootca,
		}
		ln, _ := net.Listen("tcp", hub_addr)
		ln = tls.NewListener(ln, tlsconfig)
		log.Fatal(app.Listener(ln))
	} else {
		log.Fatal(app.Listen(hub_addr))
	}
}

func (h *Hub) publishToTopic(c *fiber.Ctx) error {
	topic := c.Params("topic")
	if topic == "" {
		return c.Status(400).SendString("Please provide a topic")
	}

	var msg Message
	if err := c.BodyParser(&msg); err != nil {
		return c.Status(400).SendString("Please provide a message")
	}

	if msg.Topic == "" {
		return c.Status(400).SendString("Please provide a topic")
	}

	if msg.Data == nil {
		return c.Status(400).SendString("Please provide a message")
	}

	if err := h.Publish(msg); err != nil {
		return c.Status(500).SendString(err.Error())
	}

	return c.Status(201).SendString("ok")

}

func (h *Hub) subscribeToTopics(c *fiber.Ctx) error {
	var topics []string
	c.Set("Access-Control-Allow-Origin", "*")
	if err := c.BodyParser(&topics); err != nil {
		return c.Status(400).SendString("NOK")
	}
	id := c.Get(subscriberHeader)
	sub := h.getSubscriberFromRequest(id)
	if sub == nil {
		h.Subscribe(Subscriber{
			ID:     id,
			Topics: topics,
		})
	} else {
		for _, topic := range topics {
			if !TopicInSubscriber(topic, sub) {
				sub.Topics = append(sub.Topics, topic)
			}
		}
	}

	c.Set(subscriberHeader, id)
	return c.Status(200).SendString("OK")
}

func (h *Hub) unsubscribeTopics(c *fiber.Ctx) error {

	topic := c.Params("topic")
	id := c.Get(subscriberHeader)
	sub := h.getSubscriberFromRequest(id)
	if topic == "" {
		topics := []string{}
		if err := c.BodyParser(&topics); err != nil {
			return c.Status(400).SendString("NOK")
		}
		h.Unsubscribe(sub, topics)
	} else {
		topics := []string{topic}
		h.Unsubscribe(sub, topics)
	}
	return c.Status(200).SendString("OK")
}

func (h *Hub) getMessages(c *fiber.Ctx) error {

	id := c.Get(subscriberHeader)
	sub := h.getSubscriberFromRequest(id)
	if sub == nil {
		return c.Status(400).SendString("Subscriber not found")
	}
	c.Response().Header.SetContentType(fiber.MIMEApplicationJSON)
	c.Set("Connection", "keep-alive")
	c.Set("Access-Control-Allow-Origin", "*")
	c.Status(200)
	enc := json.NewEncoder(c.Response().BodyWriter())
	stream := h.Registry.Subscribe(ctx, sub.Topics...)
	// for {
	// 	m, err := stream.ReceiveMessage(ctx)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	msg := Message{
	// 		Topic: m.Channel,
	// 		Data:  m.Payload,
	// 	}
	// 	if err := enc.Encode(msg); err != nil {
	// 		return err
	// 	}
	// 	//c.Response().Flush()

	// 	time.Sleep(500 * time.Millisecond)
	// }
	c.Context().SetBodyStreamWriter(fasthttp.StreamWriter(func(w *bufio.Writer) {
		for {
			m, err := stream.ReceiveMessage(ctx)
			if err != nil {
				return
			}
			msg := Message{
				Topic: m.Channel,
				Data:  m.Payload,
			}
			if err := enc.Encode(msg); err != nil {
				return
			}
			w.Flush()
			time.Sleep(500 * time.Millisecond)
		}
	}))
	return nil
}

func (h *Hub) getMessageTopic(c *fiber.Ctx) error {
	topic := c.Params("topic")
	if topic == "" {
		return c.Status(400).SendString("Topic is required")
	}
	//c.Response().Header().Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	c.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	//c.Response().Header().Add("Access-Control-Allow-Origin", "*")
	c.Set("Connection", "keep-alive")
	c.Set("Access-Control-Allow-Origin", "*")
	//c.Response().WriteHeader(http.StatusOK)
	c.Status(200)
	enc := json.NewEncoder(c.Response().BodyWriter())
	stream := h.Registry.Subscribe(ctx, topic)
	// for {
	// 	m, err := stream.ReceiveMessage(ctx)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	if err := enc.Encode(m.Payload); err != nil {
	// 		return err
	// 	}
	// 	c.Response().Flush()
	// 	time.Sleep(1 * time.Second)
	// }
	c.Context().SetBodyStreamWriter(fasthttp.StreamWriter(func(w *bufio.Writer) {
		for {
			m, err := stream.ReceiveMessage(ctx)
			if err != nil {
				return
			}
			msg := Message{
				Topic: m.Channel,
				Data:  m.Payload,
			}
			if err := enc.Encode(msg); err != nil {
				return
			}
			w.Flush()
			time.Sleep(500 * time.Millisecond)
		}
	}))
	return nil
}

func (h *Hub) getSubscriber(c *fiber.Ctx) error {
	id := c.Get(subscriberHeader)
	sub := h.getSubscriberFromRequest(id)
	if sub == nil {
		return c.Status(400).SendString("Subscriber not found")
	}
	return c.Status(200).JSON(sub)
}
