package main

import (
	"crypto/x509"
	"log"
	"os"

	ptls "github.com/jim-minter/proxy/pkg/tls"
)

func run() error {
	hostname, err := os.Hostname()
	if err != nil {
		return err
	}

	if err = ptls.NewCertificate("proxy.crt", "proxy.key", hostname, x509.ExtKeyUsageServerAuth); err != nil {
		return err
	}

	if err = ptls.NewCertificate("client.crt", "client.key", "client", x509.ExtKeyUsageClientAuth); err != nil {
		return err
	}

	if err = ptls.NewCertificate("sampleserver.crt", "sampleserver.key", hostname, x509.ExtKeyUsageServerAuth); err != nil {
		return err
	}

	return nil
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
