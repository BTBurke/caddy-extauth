package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"

	"github.com/BTBurke/caddy-extauth/examples/services"
)

const use string = `
Try these examples:

Rejected request - should get 401:
curl http://127.0.0.1:8080/unauthorized

Allowed request - should get 200:
curl http://127.0.0.1:8080/ok


`

func main() {
	service, err := net.Listen("tcp", ":9000")
	if err != nil {
		log.Fatal(err)
	}
	auth, err := net.Listen("tcp", ":9001")
	if err != nil {
		log.Fatal(err)
	}

	go services.StartAuth(auth)
	go services.StartService(service)

	fmt.Printf(use)

	signalChan := make(chan os.Signal, 1)
	cleanupDone := make(chan struct{})
	signal.Notify(signalChan, os.Interrupt)
	go func() {
		<-signalChan
		fmt.Println("\nReceived an interrupt, stopping services...")
		service.Close()
		auth.Close()
		close(cleanupDone)
	}()
	<-cleanupDone
}
