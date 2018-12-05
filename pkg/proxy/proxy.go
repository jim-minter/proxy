package proxy

import (
	"crypto/tls"
	"crypto/x509"
	"io"
	"log"
	"sync"

	ptls "github.com/jim-minter/proxy/pkg/tls"
)

type closeWriter interface {
	CloseWrite() error
}

// Proxy copies data from src to dst then issues a CloseWrite or Close on dst.
// It is intended to be called on a new goroutine, hence it logs any errors
// rather than returning them.
func Proxy(wg *sync.WaitGroup, dst io.WriteCloser, src io.Reader) {
	defer wg.Done()

	if err := func() error {
		if _, err := io.Copy(dst, src); err != nil {
			return err
		}

		if dst, ok := dst.(closeWriter); ok {
			return dst.CloseWrite()
		}

		return dst.Close()
	}(); err != nil {
		log.Print(err)
	}
}

// ClientTLSConfig returns a suitable *tls.Config for a client.
func ClientTLSConfig() (*tls.Config, error) {
	proxyCert, err := ptls.ReadCertificate("proxy.crt", "")
	if err != nil {
		return nil, err
	}

	clientCert, err := ptls.ReadCertificate("client.crt", "client.key")
	if err != nil {
		return nil, err
	}

	rootCAs, err := x509.SystemCertPool()
	if err != nil {
		return nil, err
	}
	rootCAs.AddCert(proxyCert.Leaf)

	return &tls.Config{
		Certificates: []tls.Certificate{clientCert},
		RootCAs:      rootCAs,
	}, nil
}
