package services

import (
	"fmt"
	"log"
	"net"
	"net/http"
)

func success(w http.ResponseWriter, r *http.Request) {
	out := fmt.Sprintf("Success! Received request:\n%+v\n", r)
	w.Write([]byte(out))
	w.WriteHeader(http.StatusOK)
}

// StartService will start a service listening on port 8080 that echoes
// whatever request it receives to the terminal
func StartService(l net.Listener) {
	log.Printf("Service listening on port 9000\n")
	h := http.NewServeMux()
	h.HandleFunc("/", success)
	s := &http.Server{
		Addr:    ":9000",
		Handler: h,
	}
	log.Print(s.Serve(l))
}
