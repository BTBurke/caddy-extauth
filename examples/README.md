# Extauth Examples

This directory contains examples of using extauth in both the simple and router mode.  These examples use Go, but you can use any language to implement your authorization service.

# Simple Mode

In simple mode, all authorization requests come to the root of your authorization API.  Parse the URL passed in the `X-Auth-URL` header if you want to take different authorization actions based on path, query parameters, etc.

```go
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
```

# Router Mode

In router mode, authorization requests propagate the path and query parameters of the original request so that you can use a router implementation in your language to route each request to an appropriate authorization handler.  This example uses [chi](https://github.com/go-chi/chi) to set up a router that has different handlers for the `/unauthorized` and `/ok` paths.

```go
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
```

# Running the Examples

First, run caddy in the directory with the appropriate Caddyfile

```
caddy -conf Caddyfile.router
```

Then in a separate terminal, start the auth and passthrough service:

```
go run router-mode.go
```

Now try to curl the two possible paths:

```
# should respond 200 OK from the pass through service
curl http://127.0.0.1/ok

# should respond 401 UNAUTHORIZED
curl http://127.0.0.1/unauthorized
```
