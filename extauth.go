package extauth

import (
	"crypto/tls"
	"io/ioutil"
	"net/http"
	"net/url"
)

func (a *Auth) ServeHTTP(w http.ResponseWriter, r *http.Request) (int, error) {

	// create client if it doesn't exist, in normal operation client should be nil
	// but having the client as part of the auth struct is useful for testing
	if a.client == nil {
		a.client = &http.Client{}
	}
	a.client.Timeout = a.Timeout

	url, err := url.Parse(a.Proxy)
	if err != nil {
		return handleUnathorized(w, nil), nil
	}
	if url.Scheme == "https" && a.InsecureSkipVerify {
		a.client.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}
	req, err := http.NewRequest("GET", url.String(), nil)
	if err != nil {
		return handleUnathorized(w, nil), nil
	}

	if a.Cookies {
		for _, c := range r.Cookies() {
			req.AddCookie(c)
		}
	}

	if a.Headers {
		req.Header = r.Header
		// Retain original host header
		req.Host = r.Host
		req.Header.Add("X-URL", r.URL.String())
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return handleUnathorized(w, nil), nil
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		for _, c := range resp.Cookies() {
			r.AddCookie(c)
		}
		r.Header = resp.Header
		return a.Next.ServeHTTP(w, r)
	}

	respReason, err := ioutil.ReadAll(resp.Body)
	return handleUnathorized(w, respReason), nil
}

func handleUnathorized(w http.ResponseWriter, resp []byte) int {
	w.WriteHeader(http.StatusUnauthorized)
	w.Write(resp)
	return http.StatusUnauthorized
}
