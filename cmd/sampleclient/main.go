package main

import (
	"io"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/jim-minter/proxy/pkg/proxy"
	ptls "github.com/jim-minter/proxy/pkg/tls"
)

// getProxyTransport returns an http.RoundTripper suitable for talking to the
// proxy.
func getProxyTransport(proxyURL string) (http.RoundTripper, error) {
	url, err := url.Parse(proxyURL)
	if err != nil {
		return nil, err
	}

	config, err := proxy.ClientTLSConfig()
	if err != nil {
		return nil, err
	}

	return &http.Transport{
		TLSClientConfig: config,
		Proxy:           http.ProxyURL(url),
	}, nil
}

func run(proxyURL, getURL string) error {
	transport, err := getProxyTransport(proxyURL)
	if err != nil {
		return err
	}

	cert, err := ptls.ReadCertificate("sampleserver.crt", "")
	if err != nil {
		return err
	}
	transport.(*http.Transport).TLSClientConfig.RootCAs.AddCert(cert.Leaf)

	cli := http.Client{
		Transport: transport,
	}

	resp, err := cli.Get(getURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, err = io.Copy(os.Stdout, resp.Body)
	return err
}

func main() {
	if len(os.Args) != 3 {
		log.Fatalf("usage: %s proxyURL getURL\n", os.Args[0])
	}

	if err := run(os.Args[1], os.Args[2]); err != nil {
		log.Fatal(err)
	}
}
