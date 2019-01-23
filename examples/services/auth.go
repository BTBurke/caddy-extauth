package services

import (
	"log"
	"net"
	"net/http"
)

func unauthorized(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusUnauthorized)
}

func ok(w http.ResponseWriter, r *http.Request) {
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

func StartAuth(l net.Listener) {
	log.Printf("Authorization service listening on port 9001\n")
	h := http.NewServeMux()
	h.HandleFunc("/ok", ok)
	h.HandleFunc("/unauthorized", unauthorized)
	s := &http.Server{
		Addr:    ":9001",
		Handler: h,
	}
	log.Print(s.Serve(l))
}
