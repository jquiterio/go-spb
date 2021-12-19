/*
 * @file: utils.go
 * @author: Jorge Quitério
 * @copyright (c) 2021 Jorge Quitério
 * @license: MIT
 */

package utils

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net"
	"os"
)

type ClientCerts struct {
	Cert []byte
	Key  *ecdsa.PrivateKey
}

func GetIPAddr() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}

func WriteClientCertsToFile() error {
	cert, err := GetClientCerts()
	if err != nil {
		return err
	}
	certout, err := os.Create("client.pem")
	if err != nil {
		return err
	}
	defer certout.Close()
	err = pem.Encode(certout, &pem.Block{Type: "CERTIFICATE", Bytes: cert.Cert})
	if err != nil {
		return err
	}
	keyout, err := os.OpenFile("client.key", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer keyout.Close()
	err = pem.Encode(keyout, &pem.Block{Type: "EC PRIVATE KEY", Bytes: PrivateKeyToByte(cert.Key)})
	if err != nil {
		return err
	}
	return nil
}

func PrintClientCerts() {
	certs, err := GetClientCerts()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println("Client Certificate:")
	pem.Encode(os.Stdout, &pem.Block{Type: "CERTIFICATE", Bytes: certs.Cert})
	fmt.Println(string(certs.Cert))
	fmt.Println("Client Key:")
	fmt.Println(string(PrivateKeyToByte(certs.Key)))
}

func PrivateKeyToByte(key *ecdsa.PrivateKey) []byte {
	b, _ := x509.MarshalECPrivateKey(key)
	return b
}

func GetClientCerts() (ClientCerts, error) {
	certs, err := GenCert()
	if err != nil {
		return ClientCerts{}, err
	}
	return ClientCerts{Cert: certs.ClientCert, Key: certs.ClientKey}, nil
}
