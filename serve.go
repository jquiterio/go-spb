/*
 * @file: serve.go
 * @author: Jorge Quitério
 * @copyright (c) 2021 Jorge Quitério
 * @license: MIT
 */

package mhub

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type (
	SubscriberRequestConfig struct {
		Skipper      Skipper
		TargetHeader string
	}
	Skipper func(echo.Context) bool
)

var subscriberHeader string = "X-Subscriber-ID"

func HandlerSubscriberRequest() echo.MiddlewareFunc {
	sConfig := SubscriberRequestConfig{
		Skipper:      middleware.DefaultSkipper,
		TargetHeader: subscriberHeader,
	}
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if sConfig.Skipper(c) {
				return next(c)
			}
			req := c.Request()
			res := c.Response()
			sid := req.Header.Get(sConfig.TargetHeader)
			if sid == "" {
				return c.JSON(http.StatusUnauthorized, "Please provide a valid subscriber id")
			}
			res.Header().Set(sConfig.TargetHeader, sid)
			return next(c)
		}
	}
}

func (h *Hub) getSubscriberFromRequest(c echo.Context) *Subscriber {
	sub := c.Request().Header.Get(subscriberHeader)
	return h.GetSubscriber(sub)
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

func (h *Hub) Serve() {
	conf := GetFromEnvOrDefault()

	e := echo.New()
	e.Use(middleware.SecureWithConfig(middleware.SecureConfig{
		XSSProtection:         "",
		ContentTypeNosniff:    "",
		XFrameOptions:         "",
		HSTSMaxAge:            3600,
		ContentSecurityPolicy: "default-src 'self'",
	}))
	e.Use(middleware.Logger())
	e.Use(HandlerSubscriberRequest())
	// e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
	// 	Skipper:      middleware.DefaultSkipper,
	// AllowOrigins: []string{"*"},
	// AllowMethods: []string{http.MethodGet, http.MethodHead, http.MethodPut, http.MethodPatch, http.MethodPost, http.MethodDelete, http.MethodOptions},
	// })
	e.Use(middleware.CORS())

	e.GET("/", h.getMessages)
	e.GET("/me", h.getSubscriber)
	e.GET("/:topic", h.getMessageTopic)
	e.POST("subscribe", h.subscribeToTopics)
	e.POST("/unsubscribe", h.unsubscribeTopics)
	e.POST("/publish/:topic", h.publishToTopic)

	hub_addr := conf.Hub.Addr + ":" + conf.Hub.Port
	fmt.Println("Hub is listening on: " + hub_addr)
	//e.Logger.Fatal(e.Start(conf.Hub.Addr + ":" + conf.Hub.Port))

	// HTTPS
	if conf.Hub.Secure {
		caPem, err := ioutil.ReadFile("ca.pem")
		if err != nil {
			genCertError(err)
			return
		}
		rooca := x509.NewCertPool()
		if ok := rooca.AppendCertsFromPEM(caPem); !ok {
			panic("Failed to append CA cert")
		}

		s := http.Server{
			Addr:    hub_addr,
			Handler: e,
			TLSConfig: &tls.Config{
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
				ClientCAs:  rooca,
			},
			TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler), 0),
		}

		if err := s.ListenAndServeTLS("server.pem", "server.key"); err != nil {
			panic(err)
		}
	} else {
		// s := http.Server{
		// 	Addr:    hub_addr,
		// 	Handler: e,
		// }
		// if err := s.ListenAndServe(); err != nil {
		// 	panic(err)
		// }
		e.Logger.Fatal(e.Start(hub_addr))
	}
}

func (h *Hub) publishToTopic(c echo.Context) error {

	topic := c.Param("topic")
	if topic == "" {
		return c.JSON(400, echo.Map{
			"msg": "Topic is required",
		})
	}
	//sub := h.getSubscriberFromRequest(c)

	var msg Message
	if err := c.Bind(&msg); err != nil {
		return c.JSON(400, err)
	}

	if msg.Topic == "" {
		return c.JSON(400, echo.Map{
			"msg": "Topic is required",
		})
	}
	if msg.Data == nil {
		return c.JSON(400, echo.Map{
			"msg": "Data is required",
		})
	}

	if err := h.Publish(msg); err != nil {
		return c.JSON(400, echo.Map{
			"msg": err,
		})
	}
	return c.JSON(201, echo.Map{
		"msg": "OK",
	})
}

func (h *Hub) subscribeToTopics(c echo.Context) error {
	topics := []string{}
	if err := c.Bind(&topics); err != nil {
		return c.JSON(400, echo.Map{
			"msg": err,
		})
	}
	sub := h.getSubscriberFromRequest(c)
	if sub == nil {
		id := c.Request().Header.Get(subscriberHeader)
		h.Subscribe(Subscriber{
			ID:     id,
			Topics: topics,
		})
	} else {
		h.Subscribe(Subscriber{
			ID:     sub.ID,
			Topics: topics,
		})
	}
	return c.JSON(200, echo.Map{
		"msg": "Subscribed to Topics: " + fmt.Sprint(topics),
	})
}

func (h *Hub) unsubscribeTopics(c echo.Context) error {

	topic := c.Param("topic")
	sub := h.getSubscriberFromRequest(c)
	if topic == "" {
		topics := []string{}
		if err := c.Bind(&topics); err != nil {
			return c.JSON(400, echo.Map{
				"msg": err,
			})
		}
		h.Unsubscribe(sub, topics)
	} else {
		topics := []string{topic}
		h.Unsubscribe(sub, topics)
	}
	return c.JSON(200, echo.Map{
		"msg": "OK",
	})
}

func (h *Hub) getMessages(c echo.Context) error {

	sub := h.getSubscriberFromRequest(c)
	if sub == nil {
		return c.JSON(400, echo.Map{
			"msg": "Subscriber not found",
		})
	}
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	c.Response().WriteHeader(http.StatusOK)
	enc := json.NewEncoder(c.Response())
	stream := h.Registry.Subscribe(ctx, sub.Topics...)
	for {
		m, err := stream.ReceiveMessage(ctx)
		if err != nil {
			return err
		}
		msg := Message{
			Topic: m.Channel,
			Data:  m.Payload,
		}
		if err := enc.Encode(msg); err != nil {
			return err
		}
		c.Response().Flush()
		time.Sleep(1 * time.Second)
	}
}

func (h *Hub) getMessageTopic(c echo.Context) error {
	topic := c.Param("topic")
	if topic == "" {
		return c.JSON(400, echo.Map{
			"msg": "Topic is required",
		})
	}
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	c.Response().WriteHeader(http.StatusOK)
	enc := json.NewEncoder(c.Response())
	stream := h.Registry.Subscribe(ctx, topic)
	for {
		m, err := stream.ReceiveMessage(ctx)
		if err != nil {
			return err
		}
		if err := enc.Encode(m.Payload); err != nil {
			return err
		}
		c.Response().Flush()
		time.Sleep(1 * time.Second)
	}
}

func (h *Hub) getSubscriber(c echo.Context) error {
	sub := h.getSubscriberFromRequest(c)
	if sub == nil {
		return c.JSON(400, echo.Map{
			"msg": "Subscriber not found",
		})
	}
	return c.JSON(200, sub)
}
