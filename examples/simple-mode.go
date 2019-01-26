package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
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

// SimpleAuth is an example of using the simple mode to parse the URL passed in X-Auth-URL
func SimpleAuth() http.Handler {
	h := http.NewServeMux()
	h.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("got request: %+v", r.URL)

		// in simple mode, parse the header to recover the URL query parameters and path
		url, _ := url.Parse(r.Header.Get("X-Auth-URL"))

		switch url.Path {
		case "/unauthorized":
			w.WriteHeader(http.StatusUnauthorized)
		case "/ok":
			// Copy headers so they will propagate to the next middleware
			for k, v := range r.Header {
				for _, v1 := range v {
					w.Header().Add(k, v1)
				}
			}

			// Copy cookies for same reason.  Technically, this is only required if you
			// plan to set a new cookie on the client.  If you just want to pass along the
			// existing cookies, copying existing headers is enough.
			for _, cookie := range r.Cookies() {
				http.SetCookie(w, cookie)
			}
			w.WriteHeader(http.StatusOK)
		default:
			w.WriteHeader(http.StatusUnauthorized)
		}
	})
	return h
}

func main() {
	service, err := net.Listen("tcp", ":9000")
	if err != nil {
		log.Fatal(err)
	}
	auth, err := net.Listen("tcp", ":9001")
	if err != nil {
		log.Fatal(err)
	}

	// start the simple-mode auth service and a passthrough service on 9000 and 9001
	go services.StartAuth(auth, SimpleAuth())
	go services.StartService(service)

	fmt.Printf(use)

	// can ignore this, just traps signals to shut down the two services when you are done
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
