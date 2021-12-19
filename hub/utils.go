/*
 * @file: utils.go
 * @author: Jorge Quitério
 * @copyright (c) 2021 Jorge Quitério
 * @license: MIT
 */

package hub

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
)

type ClientCerts struct {
	Cert []byte
	Key  *ecdsa.PrivateKey
}

func (h HubServer) GetClientCerts() (ClientCerts, error) {
	//certs, err := GenCert()
	certs := h.Certificate
	return ClientCerts{Cert: certs.ClientCert, Key: certs.ClientKey}, nil
}

func PrivateKeyToByte(key *ecdsa.PrivateKey) []byte {
	b, _ := x509.MarshalECPrivateKey(key)
	return b
}

func (h HubServer) PrintClientCerts() {
	certs, err := h.GetClientCerts()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println("Client Certificate:")
	pem.Encode(os.Stdout, &pem.Block{Type: "CERTIFICATE", Bytes: certs.Cert})
	fmt.Println("Client Key:")
	pem.Encode(os.Stdout, &pem.Block{Type: "EC PRIVATE KEY", Bytes: PrivateKeyToByte(certs.Key)})
	//fmt.Println(string(PrivateKeyToByte(certs.Key)))
}

func (h HubServer) GetClientPrivateKey() *pem.Block {
	certs, err := h.GetClientCerts()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return &pem.Block{Type: "EC PRIVATE KEY", Bytes: PrivateKeyToByte(certs.Key)}
}

func (h HubServer) GetClientCert() *pem.Block {
	certs, err := h.GetClientCerts()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return &pem.Block{Type: "CERTIFICATE", Bytes: certs.Cert}
}

func (h HubServer) GetServerPrivateKey() *pem.Block {
	certs := h.Certificate
	return &pem.Block{Type: "EC PRIVATE KEY", Bytes: PrivateKeyToByte(certs.CertKey)}
}

func (h HubServer) GetServerCert() *pem.Block {
	certs := h.Certificate
	return &pem.Block{Type: "CERTIFICATE", Bytes: certs.Cert}
}

func (h HubServer) WriteClientCertsToFile(certType string) error {
	basepath := "certs/"
	certs := h.Certificate
	var filename string
	var cert []byte
	var key *ecdsa.PrivateKey
	switch certType {
	case "client":
		filename = "client"
		cert = certs.ClientCert
		key = certs.ClientKey
	case "server":
		filename = "server"
		cert = certs.Cert
		key = certs.CertKey
	case "ca":
		filename = "ca"
		cert = certs.RootCA
	default:
		return fmt.Errorf("Invalid certType: %s", certType)
	}
	certout, err := os.Create(basepath + filename + ".pem")
	if err != nil {
		return err
	}
	defer certout.Close()
	err = pem.Encode(certout, &pem.Block{Type: "CERTIFICATE", Bytes: cert})
	if err != nil {
		return err
	}
	if certType != "ca" {
		keyout, err := os.OpenFile(basepath+filename+".key", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
		if err != nil {
			return err
		}
		defer keyout.Close()
		err = pem.Encode(keyout, &pem.Block{Type: "EC PRIVATE KEY", Bytes: PrivateKeyToByte(key)})
		if err != nil {
			return err
		}
	}
	return nil
}
