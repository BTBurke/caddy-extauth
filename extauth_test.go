package extauth

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/mholt/caddy/caddyhttp/httpserver"
	"github.com/stretchr/testify/assert"
)

func passThruHandler(w http.ResponseWriter, r *http.Request) (int, error) {
	return http.StatusOK, nil
}

func authorizedHandler(w http.ResponseWriter, r *http.Request) {
	// copy received headers and cookies back into response so they can be inspected
	for head, val := range r.Header {
		for _, val1 := range val {
			w.Header().Add(head, val1)
		}
	}
	for _, c := range r.Cookies() {
		http.SetCookie(w, c)
	}
	w.WriteHeader(http.StatusOK)
}

func forbiddenHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusUnauthorized)
	w.Write([]byte("you are not authorized"))
}

func serverErrorHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusInternalServerError)
}

func timeoutHandler(w http.ResponseWriter, r *http.Request) {
	time.Sleep(3 * time.Second)
	w.WriteHeader(http.StatusOK)
}

func TestAuthorization(t *testing.T) {

	tests := []struct {
		authServerHandler http.HandlerFunc
		status            int
	}{
		{forbiddenHandler, http.StatusUnauthorized},
		{authorizedHandler, http.StatusOK},
		{serverErrorHandler, http.StatusUnauthorized},
		//{timeoutHandler, http.StatusUnauthorized},
	}

	for _, test := range tests {
		ts := httptest.NewServer(test.authServerHandler)
		defer ts.Close()
		auth := &Auth{
			Proxy:   ts.URL,
			Next:    httpserver.HandlerFunc(passThruHandler),
			Timeout: time.Duration(0 * time.Second),
			client:  ts.Client(),
		}
		req := httptest.NewRequest("GET", "http://protected.local", nil)
		w := httptest.NewRecorder()
		status, err := auth.ServeHTTP(w, req)
		resp := w.Result()
		assert.NoError(t, err)
		assert.Equal(t, test.status, status)
		assert.Equal(t, test.status, resp.StatusCode)
	}
}
