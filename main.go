/*
 * @file: main.go
 * @author: Jorge Quitério
 * @copyright (c) 2021 Jorge Quitério
 * @license: MIT
 */

package main

import "github.com/jquiterio/go-spb/hub"

func main() {

	h := hub.NewHub()
	h.Serve()

}
