// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Modified 2016 by Steve Manuel, Boss Sauce Creative, LLC
// All modifications are relicensed under the same BSD license
// found in the LICENSE file.

// Generate a self-signed X.509 certificate for a TLS server. Outputs to
// 'devcerts/cert.pem' and 'devcerts/key.pem' and will overwrite existing files.

package tls

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"log"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/ponzu-cms/ponzu/system/db"
)

func publicKey(priv interface{}) interface{} {
	switch k := priv.(type) {
	case *rsa.PrivateKey:
		return &k.PublicKey
	case *ecdsa.PrivateKey:
		return &k.PublicKey
	default:
		return nil
	}
}

func pemBlockForKey(priv interface{}) *pem.Block {
	switch k := priv.(type) {
	case *rsa.PrivateKey:
		return &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(k)}
	case *ecdsa.PrivateKey:
		b, err := x509.MarshalECPrivateKey(k)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to marshal ECDSA private key: %v", err)
			os.Exit(2)
		}
		return &pem.Block{Type: "EC PRIVATE KEY", Bytes: b}
	default:
		return nil
	}
}

func setupDev() {
	var priv interface{}
	var err error

	priv, err = rsa.GenerateKey(rand.Reader, 2048)

	if err != nil {
		log.Fatalf("failed to generate private key: %s", err)
	}

	notBefore := time.Now()
	notAfter := notBefore.Add(time.Hour * 24 * 30) // valid for 30 days

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		log.Fatalf("failed to generate serial number: %s", err)
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Ponzu Dev Server"},
		},
		NotBefore: notBefore,
		NotAfter:  notAfter,

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	hosts := []string{"localhost", "0.0.0.0"}
	domain := db.ConfigCache("domain").(string)
	if domain != "" {
		hosts = append(hosts, domain)
	}

	for _, h := range hosts {
		if ip := net.ParseIP(h); ip != nil {
			template.IPAddresses = append(template.IPAddresses, ip)
		} else {
			template.DNSNames = append(template.DNSNames, h)
		}
	}

	// make all certs CA
	template.IsCA = true
	template.KeyUsage |= x509.KeyUsageCertSign

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, publicKey(priv), priv)
	if err != nil {
		log.Fatalln("Failed to create certificate:", err)
	}

	// overwrite/create directory for devcerts
	pwd, err := os.Getwd()
	if err != nil {
		log.Fatalln("Couldn't find working directory to locate or save dev certificates:", err)
	}

	vendorTLSPath := filepath.Join(pwd, "cmd", "ponzu", "vendor", "github.com", "ponzu-cms", "ponzu", "system", "tls")
	devcertsPath := filepath.Join(vendorTLSPath, "devcerts")

	// clear all old certs if found
	err = os.RemoveAll(devcertsPath)
	if err != nil {
		log.Fatalln("Failed to remove old files from dev certificate directory:", err)
	}

	err = os.Mkdir(devcertsPath, os.ModeDir|os.ModePerm)
	if err != nil {
		log.Fatalln("Failed to create directory to locate or save dev certificates:", err)
	}

	certOut, err := os.Create(filepath.Join(devcertsPath, "cert.pem"))
	if err != nil {
		log.Fatalln("Failed to open devcerts/cert.pem for writing:", err)
	}
	pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	certOut.Close()

	keyOut, err := os.OpenFile(filepath.Join(devcertsPath, "key.pem"), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalln("Failed to open devcerts/key.pem for writing:", err)
		return
	}
	pem.Encode(keyOut, pemBlockForKey(priv))
	keyOut.Close()
}
