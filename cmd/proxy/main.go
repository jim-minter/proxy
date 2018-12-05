package main

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/jim-minter/proxy/pkg/apachelog"
	"github.com/jim-minter/proxy/pkg/proxy"
	ptls "github.com/jim-minter/proxy/pkg/tls"
)

var (
	listen  = flag.String("listen", "8443", "listen port")
	allowed = flag.String("allowed", "22,8444", "allowed proxy ports")
)

func hasPort(allowedPorts []uint16, port uint16) bool {
	for _, p := range allowedPorts {
		if p == port {
			return true
		}
	}
	return false
}

type handler struct {
	hostname     string
	allowedPorts []uint16
}

var _ http.Handler = &handler{}

func newHandler(allowedPorts []uint16) (http.Handler, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}

	return &handler{hostname: hostname, allowedPorts: allowedPorts}, nil
}

func (h *handler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	hijacker, ok := rw.(http.Hijacker)
	if !ok {
		http.Error(rw, "ResponseWriter is not a Hijacker", http.StatusInternalServerError)
		return
	}

	if req.Method != http.MethodConnect {
		http.Error(rw, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	if req.URL.Hostname() != h.hostname {
		http.Error(rw, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		return
	}

	port, err := strconv.ParseUint(strings.TrimSpace(req.URL.Port()), 10, 16)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	if !hasPort(h.allowedPorts, uint16(port)) {
		http.Error(rw, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		return
	}

	serverConn, err := net.Dial("tcp", req.URL.Host)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}
	defer serverConn.Close()

	clientConn, buf, err := hijacker.Hijack()
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	defer clientConn.Close()

	_, err = clientConn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
	if err != nil {
		log.Print(err)
		return
	}

	_, err = io.CopyN(serverConn, buf, int64(buf.Reader.Buffered()))
	if err != nil {
		log.Print(err)
		return
	}

	wg := &sync.WaitGroup{}
	wg.Add(2)

	go proxy.Proxy(wg, clientConn.(*tls.Conn), serverConn)
	go proxy.Proxy(wg, serverConn.(*net.TCPConn), clientConn)

	wg.Wait()
}

func run(listenPort uint16, allowedPorts []uint16) error {
	h, err := newHandler(allowedPorts)
	if err != nil {
		return err
	}

	proxyCert, err := ptls.ReadCertificate("proxy.crt", "proxy.key")
	if err != nil {
		return err
	}

	clientCert, err := ptls.ReadCertificate("client.crt", "")
	if err != nil {
		return err
	}

	clientCAs := x509.NewCertPool()
	clientCAs.AddCert(clientCert.Leaf)

	config := &tls.Config{
		Certificates: []tls.Certificate{proxyCert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    clientCAs,
	}

	l, err := tls.Listen("tcp", fmt.Sprintf(":%d", listenPort), config)
	if err != nil {
		return err
	}

	log.Print("proxy is running")

	return http.Serve(l, apachelog.NewHandler(h))
}

func main() {
	flag.Parse()

	listenPort, err := strconv.ParseUint(strings.TrimSpace(*listen), 10, 16)
	if err != nil {
		log.Fatal(err)
	}

	var allowedPorts []uint16
	for _, port := range strings.Split(*allowed, ",") {
		p, err := strconv.ParseUint(strings.TrimSpace(port), 10, 16)
		if err != nil {
			log.Fatal(err)
		}
		allowedPorts = append(allowedPorts, uint16(p))
	}

	if err := run(uint16(listenPort), allowedPorts); err != nil {
		log.Fatal(err)
	}
}
