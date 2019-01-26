package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"

	"github.com/BTBurke/caddy-extauth/examples/services"
	"github.com/go-chi/chi"
)

const use string = `
Try these examples:

Rejected request - should get 401:
curl http://127.0.0.1:8080/unauthorized

Allowed request - should get 200:
curl http://127.0.0.1:8080/ok


`

// RouterAuth is an example of using the Chi router to handle incoming requests using path parameters
// to route each request to the appropriate authorization handler
func RouterAuth() *chi.Mux {
	r := chi.NewRouter()
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Request: %+v\n", r)
		w.WriteHeader(http.StatusUnauthorized)
	})
	r.Get("/unauthorized", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	})
	r.Get("/ok", func(w http.ResponseWriter, r *http.Request) {
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
	})
	return r
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

	// start the router-mode auth service and a passthrough service on 9000 and 9001
	go services.StartAuth(auth, RouterAuth())
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
