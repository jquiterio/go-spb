/*
 * @file: msg.go
 * @author: Jorge Quitério
 * @copyright (c) 2022 Jorge Quitério
 * @license: MIT
 */

package hub

type Message struct {
	subscriberID string      `json:"subscriber_id"`
	ID           string      `json:"id"`
	Topic        string      `json:"topic"`
	Type         string      `json:"type"`
	Data         interface{} `json:"msg"`
}
