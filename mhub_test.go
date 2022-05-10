/*
 * @file: subscribe_test.go
 * @author: Jorge Quitério
 * @copyright (c) 2022 Jorge Quitério
 * @license: MIT
 */

package mhub

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
)

var subscriberID string = "c0e74fc4-97f4-4341-91fc-e3db42b9ceaa"
var app *fiber.App

func init() {

	hub := NewHub()
	hub.Subscribe(Subscriber{
		ID: subscriberID,
	})
	app = hub.newServer()
}

func TestSubscriberRequest(t *testing.T) {
	body := []byte(`["topic1", "topic2"]`)
	req := httptest.NewRequest("POST", "/subscribe", bytes.NewReader(body))
	req.Header.Set("X-Subscriber-ID", subscriberID)
	req.Header.Add("Content-Type", "application/json")
	resp, _ := app.Test(req)
	if resp.StatusCode != 200 {
		t.Errorf("Expected status code to be 200, got %d", resp.StatusCode)
		b, _ := ioutil.ReadAll(resp.Body)
		t.Logf("response: %s", string(b))
	}
	req.Body.Close()
}

func TestPublish(t *testing.T) {
	body := map[string]interface{}{
		"subscriber_id": subscriberID,
		"topic":         "topic1",
		"type":          "publish",
		"data":          "test1",
	}
	json_body, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/publish/topic1", bytes.NewReader(json_body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Add(subscriberHeader, subscriberID)
	resp, _ := app.Test(req)
	if resp.StatusCode != 201 {
		t.Errorf("Expected status code to be 201, got %d", resp.StatusCode)
		b, _ := ioutil.ReadAll(resp.Body)
		t.Logf("response: %s", string(b))
	}
}

// func TestGetMessages(t *testing.T) {
// 	req := httptest.NewRequest("GET", "/", nil)
// 	req.Header.Set("X-Subscriber-ID", subscriberID)
// 	resp, _ := app.Test(req)

// }
