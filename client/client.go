/*
 * @file: client.go
 * @author: Jorge Quitério
 * @copyright (c) 2021 Jorge Quitério
 * @license: MIT
 */

package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/gofrs/uuid"
	"github.com/golang/glog"
)

type Client struct {
	SubscriberID string
	Topics       []string
	HubAddr      string
	conn         http.Client
	req          http.Request
	resp         http.Response
}

func NewHubClient(address string) (*Client, error) {
	a, err := url.Parse("http://" + address)
	if err != nil {
		glog.Fatal("Hub address must be a valid URL")
		return nil, err
	}
	return &Client{
		HubAddr: a.String(),
	}, nil
}

func (c *Client) AddTopic(topic []string) (ok bool) {
	if len(topic) == 0 {
		glog.Error("no topics to add")
		return
	}
	c.Topics = append(c.Topics, topic...)
	return true
}

func (c *Client) Subscribe() (ok bool) {
	url := fmt.Sprintf("%s/subscribe", c.HubAddr)
	body := []byte(``)
	if len(c.Topics) == 0 {
		glog.Error("no topics to subscribe")
		return false
	} else if len(c.Topics) > 1 {
		url = fmt.Sprintf("%s/subscribe", c.HubAddr)
		body, _ = json.Marshal(c.Topics)
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		glog.Fatal(err)
		return
	}
	req.Header.Set("X-Subscriber-ID", c.SubscriberID)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		glog.Fatal(err)
		return
	}
	if resp.StatusCode != http.StatusCreated {
		glog.Fatal("unexpected status code: ", resp.StatusCode)
	}
	return true
}

func (c *Client) Unsubscribe(topics []string) (ok bool) {
	var url string
	if len(topics) == 0 {
		glog.Error("no topics to unsubscribe")
		return
	}
	if len(topics) > 1 {
		url = fmt.Sprintf("%s/unsubscribe", c.HubAddr)
	} else {
		url = fmt.Sprintf("%s/unsubscribe/%s", c.HubAddr, topics[0])
	}
	body, _ := json.Marshal(topics)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		glog.Fatal(err)
	}
	req.Header.Set("X-Subscriber-ID", c.SubscriberID)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		glog.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		glog.Fatal("unexpected status code: ", resp.StatusCode)
	}
	return true
}

func (c *Client) Publish(topic string, msg interface{}) {
	url := fmt.Sprintf("%s/publish", c.HubAddr)
	body, err := json.Marshal(map[string]interface{}{
		"topic":    topic,
		"msg_type": "publish",
		"msg":      msg,
		"msg_id":   uuid.Must(uuid.NewV4()).String(),
	})
	if err != nil {
		glog.Fatal(err)
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		glog.Fatal(err)
	}
	req.Header.Set("X-Subscriber-ID", c.SubscriberID)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		glog.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		glog.Fatal("unexpected status code: ", resp.StatusCode)
	}
}

func (c *Client) GetMessages() {
	url := c.HubAddr
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		glog.Fatal(err)
	}
	req.Header.Set("X-Subscriber-ID", c.SubscriberID)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		glog.Fatal(err)
	}
	dec := json.NewDecoder(resp.Body)
	for {
		var message interface{}
		err := dec.Decode(&message)
		if err != nil {
			if err == io.EOF {
				break
			}
			glog.Fatal(err)
		}
		glog.Infof("Got Mesage: %+v", message)
	}
}

func (c *Client) GetTopicMessage(topic string) {
	url := c.HubAddr + "/" + topic
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		glog.Fatal(err)
	}
	req.Header.Set("X-Subscriber-ID", c.SubscriberID)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		glog.Fatal(err)
	}
	dec := json.NewDecoder(resp.Body)
	for {
		var message interface{}
		err := dec.Decode(&message)
		if err != nil {
			if err == io.EOF {
				break
			}
			glog.Fatal(err)
		}
		glog.Infof("Got Mesage: %+v", message)
	}
}
