package services

import (
	"log"
	"net"
	"net/http"
	"net/url"
)

func handleAuth(w http.ResponseWriter, r *http.Request) {
	log.Printf("got request: %+v", r.URL)

	url, _ := url.Parse(r.Header.Get("X-URL"))
	switch url.Path {
	case "/unauthorized":
		w.WriteHeader(http.StatusUnauthorized)
	default:
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
	}
}

func StartAuth(l net.Listener) {
	log.Printf("Authorization service listening on port 9001\n")
	h := http.NewServeMux()
	h.HandleFunc("/", handleAuth)
	s := &http.Server{
		Addr:    ":9001",
		Handler: h,
	}
	log.Print(s.Serve(l))
}
