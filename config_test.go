package extauth

import (
	"net/http"
	"testing"
	"time"

	"github.com/caddyserver/caddy"
	"github.com/caddyserver/caddy/caddyhttp/httpserver"
	"github.com/stretchr/testify/assert"
)

var EmptyNext = httpserver.HandlerFunc(func(w http.ResponseWriter, r *http.Request) (int, error) {
	return 0, nil
})

func TestBasicParse(t *testing.T) {
	c := caddy.NewTestController("http", `extauth https://testserver:9000`)
	err := Setup(c)
	assert.NoError(t, err)
}

func TestParsing(t *testing.T) {
	tests := []struct {
		input     string
		shouldErr bool
		expect    Auth
	}{
		{"extauth https://testserver:9000", false, Auth{Proxy: "https://testserver:9000", Headers: true, Cookies: true, Timeout: time.Duration(30 * time.Second)}},
		{"extauth {\nproxy https://testserver:9000\n}", false, Auth{Proxy: "https://testserver:9000", Headers: true, Cookies: true, Timeout: time.Duration(30 * time.Second)}},
		{"extauth {\nproxy testserver:9000\n}", false, Auth{Proxy: "http://testserver:9000", Headers: true, Cookies: true, Timeout: time.Duration(30 * time.Second)}},
		{"extauth", true, Auth{}},
		{"extauth {\nproxy https://testserver:9000\ncookies false\nheaders false\n}", false, Auth{Proxy: "https://testserver:9000", Headers: false, Cookies: false, Timeout: time.Duration(30 * time.Second)}},
		{"extauth {\nproxy https://testserver:9000\ncookies false\nheaders false\ntimeout 60s\ninsecure_skip_verify\nrouter\n}", false, Auth{Proxy: "https://testserver:9000", Router: true, Headers: false, Cookies: false, Timeout: time.Duration(60 * time.Second), InsecureSkipVerify: true}},
	}
	for _, test := range tests {
		c := caddy.NewTestController("http", test.input)
		actual, err := parse(c)
		if test.shouldErr {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
			assert.Equal(t, test.expect.Proxy, actual.Proxy)
			assert.Equal(t, test.expect.Cookies, actual.Cookies)
			assert.Equal(t, test.expect.Headers, actual.Headers)
		}
	}
}
