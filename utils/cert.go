/*
 * @file: cert.go
 * @author: Jorge Quitério
 * @copyright (c) 2021 Jorge Quitério
 * @license: MIT
 */

package utils

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"net"
	"os"
	"time"
)

type AuthCert struct {
	RootCA     []byte
	Cert       []byte
	ClientCert []byte
	CertKey    *ecdsa.PrivateKey
	ClientKey  *ecdsa.PrivateKey
}

func GenCert() (auth AuthCert, err error) {
	host, err := os.Hostname()
	if err != nil {
		return
	}
	notBefore := time.Now()
	notAfter := notBefore.Add(365 * 24 * time.Hour)
	snl := new(big.Int).Lsh(big.NewInt(1), 128)
	sn, _ := rand.Int(rand.Reader, snl)
	rootKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	rootTemp := x509.Certificate{
		SerialNumber: sn,
		Subject: pkix.Name{
			Organization: []string{host},
			CommonName:   "Root CA",
		},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	rootCAcert, err := x509.CreateCertificate(rand.Reader, &rootTemp, &rootTemp, &rootKey.PublicKey, rootKey)
	if err != nil {
		return
	}
	certKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	sn, _ = rand.Int(rand.Reader, snl)
	certTemp := x509.Certificate{
		SerialNumber: sn,
		Subject: pkix.Name{
			Organization: []string{host},
			CommonName:   "Certificate",
		},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  false,
	}
	hosts := []string{GetIPAddr()}
	for _, h := range hosts {
		if ip := net.ParseIP(h); ip != nil {
			certTemp.IPAddresses = append(certTemp.IPAddresses, ip)
		} else {
			certTemp.DNSNames = append(certTemp.DNSNames, h)
		}
	}
	cert, err := x509.CreateCertificate(rand.Reader, &certTemp, &rootTemp, &certKey.PublicKey, rootKey)
	if err != nil {
		return
	}
	clientKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	clientCertTemp := x509.Certificate{
		SerialNumber: new(big.Int).SetInt64(4),
		Subject: pkix.Name{
			Organization: []string{host},
			CommonName:   "ClientCert",
		},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
		IsCA:                  false,
	}
	clientCert, err := x509.CreateCertificate(rand.Reader, &clientCertTemp, &rootTemp, &clientKey.PublicKey, rootKey)
	if err != nil {
		return
	}
	return AuthCert{
		RootCA:     rootCAcert,
		Cert:       cert,
		ClientCert: clientCert,
		CertKey:    certKey,
		ClientKey:  clientKey,
	}, nil
}
