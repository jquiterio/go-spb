/*
 * @file: serve.go
 * @author: Jorge Quitério
 * @copyright (c) 2021 Jorge Quitério
 * @license: MIT
 */

package hub

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/jquiterio/go-spb/config"
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
	config := SubscriberRequestConfig{
		Skipper:      middleware.DefaultSkipper,
		TargetHeader: subscriberHeader,
	}
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if config.Skipper(c) {
				return next(c)
			}
			req := c.Request()
			res := c.Response()
			sid := req.Header.Get(config.TargetHeader)
			if sid == "" {
				return c.JSON(http.StatusUnauthorized, "Please provide a valid subscriber id")
			}
			res.Header().Set(config.TargetHeader, sid)
			return next(c)
		}
	}
}

func (h *Hub) getSubscriberFromRequest(c echo.Context) *Subscriber {
	sub := c.Request().Header.Get(subscriberHeader)
	return h.GetSubscriber(sub)
}

func (h *Hub) Serve() {
	conf := config.Config

	e := echo.New()
	// e.Use(middleware.SecureWithConfig(middleware.SecureConfig{
	// 	XSSProtection:         "",
	// 	ContentTypeNosniff:    "",
	// 	XFrameOptions:         "",
	// 	HSTSMaxAge:            3600,
	// 	ContentSecurityPolicy: "default-src 'self'",
	// }))
	e.Use(middleware.Logger())
	e.Use(HandlerSubscriberRequest())

	e.GET("/", h.getMessages)
	e.GET("/me", h.getSubscriber)
	//e.GET("/:topic", h.getMessages)
	e.POST("subscribe", h.subscribeToTopic)
	e.POST("/unsubscribe/:topic", h.unsubscribeTopic)
	e.POST("/publish/:topic", h.publishToTopic)

	hub_addr := conf.Hub.Addr + ":" + conf.Hub.Port
	fmt.Println("Hub is listening on: " + hub_addr)
	e.Logger.Fatal(e.Start(conf.Hub.Addr + ":" + conf.Hub.Port))
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

	if err := h.Publish(msg); err != nil {
		return c.JSON(400, echo.Map{
			"msg": err,
		})
	}
	return c.JSON(201, echo.Map{
		"msg": "OK",
	})
}

func (h *Hub) subscribeToTopic(c echo.Context) error {
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

func (h *Hub) unsubscribeTopic(c echo.Context) error {

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

	//topic := c.Param("topic")
	//sub_id := c.Request().Header.Get(subscriberHeader)
	sub := h.getSubscriberFromRequest(c)
	if sub == nil {
		return c.JSON(400, echo.Map{
			"msg": "Subscriber not found",
		})
	}
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	c.Response().WriteHeader(http.StatusOK)
	enc := json.NewEncoder(c.Response())
	// for {
	// 	for _, s := range h.Subscribers {
	// 		if s.ID == sub.ID {
	// 			for _, t := range s.Topics {
	// 				stream := h.Registry.Subscribe(ctx, t)
	// 				m, err := stream.ReceiveMessage(ctx)
	// 				if err != nil {
	// 					return err
	// 				}
	// 				if t == m.Channel {
	// 					if err := enc.Encode(m.Payload); err != nil {
	// 						return err
	// 					}
	// 					c.Response().Flush()
	// 				}
	// 			}
	// 		}
	// 	}
	// 	time.Sleep(1 * time.Second)
	// }
	stream := h.Registry.Subscribe(ctx, sub.Topics...)
	//var message interface{}
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
