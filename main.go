/*
 * @file: main.go
 * @author: Jorge Quitério
 * @copyright (c) 2021 Jorge Quitério
 * @license: MIT
 */

package main

import (
	"github.com/jquiterio/go-spb/hub"
)

func main() {

	hub := hub.NewHub()
	hub.Server.PrintClientCerts()
	hub.Server.WriteClientCertsToFile("server")
	hub.Server.WriteClientCertsToFile("client")
	hub.Server.WriteClientCertsToFile("ca")
	hub.Serve()
}
