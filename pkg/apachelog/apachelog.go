package apachelog

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
	"time"
)

type handler struct {
	http.Handler
}

// NewHandler returns a wrapping http.Handler which outputs logs to stdout in
// Apache Common Log format.
func NewHandler(h http.Handler) http.Handler {
	return handler{Handler: h}
}

func (h handler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	t := time.Now()

	rw := &responseWriter{ResponseWriter: w}
	h.Handler.ServeHTTP(rw, req)

	cn := "-"
	if len(req.TLS.PeerCertificates) == 1 {
		cn = req.TLS.PeerCertificates[0].Subject.CommonName
	}

	if rw.status == 0 {
		rw.status = http.StatusOK
	}

	fmt.Printf("%s - %s [%s] \"%s %s %s\" %d %d\n",
		req.RemoteAddr,
		cn,
		t.Format("02/Jan/2006 03:04:05"),
		req.Method,
		req.RequestURI,
		req.Proto,
		rw.status,
		rw.n,
	)
}

type responseWriter struct {
	http.ResponseWriter
	n      int64
	status int
}

var _ http.Hijacker = &responseWriter{}

func (rw *responseWriter) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.n += int64(n)
	return n, err
}

func (rw *responseWriter) WriteHeader(status int) {
	rw.status = status
	rw.ResponseWriter.WriteHeader(status)
}

func (rw *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return rw.ResponseWriter.(http.Hijacker).Hijack()
}
