package extauth

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/mholt/caddy/caddyhttp/httpserver"
	"github.com/stretchr/testify/assert"
)

func makePassThru(c []*http.Cookie, h map[string][]string) httpserver.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) (int, error) {

		if c != nil {
			for _, cookie := range r.Cookies() {
				c = append(c, cookie)
			}
		}
		if h != nil {
			for k, v := range r.Header {
				h[k] = v
			}
		}
		return http.StatusOK, nil
	}
}

var host string
var urlR *url.URL

func authorizedHandler(w http.ResponseWriter, r *http.Request) {
	// check host header on proxy to extauth service
	host = r.Host
	urlR = r.URL
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
			Proxy:  ts.URL,
			Next:   httpserver.HandlerFunc(makePassThru(nil, nil)),
			client: ts.Client(),
		}
		req := httptest.NewRequest("GET", "http://protected.local", nil)
		w := httptest.NewRecorder()
		status, err := auth.ServeHTTP(w, req)
		assert.NoError(t, err)
		assert.Equal(t, test.status, status)
	}
}

func TestHeaderPassThru(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(authorizedHandler))
	defer ts.Close()

	h := make(map[string][]string)
	c := []*http.Cookie{}
	auth := &Auth{
		Proxy:   ts.URL,
		Cookies: true,
		Headers: true,
		Next:    httpserver.HandlerFunc(makePassThru(c, h)),
		client:  ts.Client(),
	}
	req := httptest.NewRequest("GET", "http://protected.local", nil)
	testCookie := &http.Cookie{
		Name:  "test",
		Value: "testing",
	}
	req.AddCookie(testCookie)
	req.Header.Add("Test-Header", "Testing")

	w := httptest.NewRecorder()
	status, err := auth.ServeHTTP(w, req)
	assert.NoError(t, err)
	assert.Equal(t, status, http.StatusOK)
	assert.Equal(t, "test=testing", h["Cookie"][0])
	assert.Equal(t, "protected.local", host)
	assert.Equal(t, "Testing", h["Test-Header"][0])
}

func TestURLPassThru(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(authorizedHandler))
	defer ts.Close()

	auth := &Auth{
		Proxy:   ts.URL + "/path/1/2/3?test=testing",
		Cookies: true,
		Headers: true,
		Next:    httpserver.HandlerFunc(makePassThru(nil, nil)),
		client:  ts.Client(),
	}
	req := httptest.NewRequest("GET", "http://protected.local/path/1/2/3?test=testing", nil)

	w := httptest.NewRecorder()
	status, err := auth.ServeHTTP(w, req)
	assert.NoError(t, err)
	assert.Equal(t, status, http.StatusOK)

	assert.NoError(t, err)
	assert.Equal(t, "protected.local", host)
	assert.Equal(t, "testing", urlR.Query().Get("test"))
	assert.Equal(t, "/path/1/2/3", urlR.Path)
}
