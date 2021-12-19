/*
 * @file: client.go
 * @author: Jorge Quitério
 * @copyright (c) 2021 Jorge Quitério
 * @license: MIT
 */

package examples

import (
	"github.com/gofrs/uuid"
	"github.com/jquiterio/go-spb/client"
)

func main() {
	id := uuid.Must(uuid.NewV4()).String()
	topics := []string{"test"}
	client := client.NewHubClient(id, topics)
	client.NewTLSConnection("localhost:8083", "client.pem", "client.key")
	subsc := client.NewSubscription("test")
	client.Subscribe(subsc)
	go client.ReceiveMsg()
}
