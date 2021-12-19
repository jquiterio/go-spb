/*
 * @file: auth_server.go
 * @author: Jorge Quitério
 * @copyright (c) 2021 Jorge Quitério
 * @license: MIT
 */

package hub

import (
	"bufio"
	"bytes"
	"encoding/gob"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/jquiterio/go-spb/utils"
)

var Cert utils.AuthCert

type ClientCert struct {
	ClientCert []byte
	ClientKey  []byte
}

func AuthServer() {

	username := os.Getenv("AUTH_USERNAME")
	password := os.Getenv("AUTH_PASSWORD")
	ip := utils.GetIPAddr()
	lst, err := net.Listen("tcp", ip+":8080")
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	defer lst.Close()
	c, err := lst.Accept()
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	buf, err := bufio.NewReader(c).ReadString('\n')
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	if strings.TrimSpace(buf) == username+":"+password {
		Cert, err := utils.GenCert()
		if err != nil {
			fmt.Println(err)
			panic(err)
		}
		ccBytes := new(bytes.Buffer)
		enc := gob.NewDecoder(ccBytes)
		err = enc.Decode(ClientCert{
			ClientCert: Cert.ClientCert,
			ClientKey:  Cert.ClientKey.X.Bytes(),
		})
		if err != nil {
			fmt.Println(err)
			panic(err)
		}
		c.Write(ccBytes.Bytes())
		lst.Close()
	} else if strings.TrimSpace(buf) == "OK" {
		lst.Close()
		return
	} else {
		c.Write([]byte("creadentials not valid"))
	}

}
