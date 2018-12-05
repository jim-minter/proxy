package main

import (
	"log"
	"net/http"
)

func run() error {
	return http.ListenAndServeTLS(":8444", "sampleserver.crt", "sampleserver.key",
		http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			rw.Write([]byte("Hello, world!\n"))
		}),
	)
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
