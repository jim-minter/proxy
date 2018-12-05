package main

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"sync"

	"github.com/jim-minter/proxy/pkg/proxy"
)

// after https://web.archive.org/web/20170510154150/http://agroman.net/corkscrew/

func run(proxyURL, sshEndpoint string) error {
	url, err := url.Parse(proxyURL)
	if err != nil {
		return err
	}

	config, err := proxy.ClientTLSConfig()
	if err != nil {
		return err
	}

	c, err := tls.Dial("tcp", url.Host, config)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(c, "CONNECT %s HTTP/1.1\r\n\r\n", sshEndpoint)
	if err != nil {
		return err
	}

	buf := bufio.NewReader(c)
	line, err := buf.ReadString('\n')
	if err != nil {
		return err
	}
	if line != "HTTP/1.1 200 OK\r\n" {
		return fmt.Errorf("unexpected response %q", line)
	}

	for line != "\r\n" {
		line, err = buf.ReadString('\n')
		if err != nil {
			return err
		}
	}

	_, err = io.CopyN(os.Stdout, buf, int64(buf.Buffered()))
	if err != nil {
		return err
	}

	wg := &sync.WaitGroup{}
	wg.Add(2)

	go proxy.Proxy(wg, os.Stdout, c)
	go proxy.Proxy(wg, c, os.Stdin)

	wg.Wait()

	return nil
}

func main() {
	if len(os.Args) != 3 {
		log.Fatalf("usage: %s proxyURL sshEndpoint\n", os.Args[0])
	}

	if err := run(os.Args[1], os.Args[2]); err != nil {
		log.Fatal(err)
	}
}
