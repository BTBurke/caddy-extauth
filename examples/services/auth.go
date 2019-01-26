package services

import (
	"log"
	"net"
	"net/http"
)

func StartAuth(l net.Listener, handler http.Handler) {
	log.Printf("Authorization service listening on port 9001\n")
	s := &http.Server{
		Addr:    ":9001",
		Handler: handler,
	}
	log.Print(s.Serve(l))
}
