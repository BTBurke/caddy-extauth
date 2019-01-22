## JWT

[![Build Status](https://travis-ci.org/BTBurke/caddy-extauth.svg?branch=master)](https://travis-ci.org/BTBurke/caddy-extauth)

**External Authorization Middleware for Caddy**

This middleware implements an authorization layer for [Caddy](https://caddyserver.com) by using an external service you provide to handle authorization for each request.  Unlike other authorization methods like JWT, this middleware provides you with more flexibility to implement your own authorization scheme.

### How it works

Every time Caddy receives a request, it will forward the request to your authorization service.  If you return a `200 OK` response, the request will continue.  Any other status code results in a `401 unauthorized`.  In your external auth service, you can do anything you want - check session cookies, validate tokens, interact with your database, etc.  This provides infinite flexibility to implement your own authorization layer and still use the rest of Caddy's features.

### Basic syntax

```
extauth [service url]
```

For example:

```
extauth https://localhost:9898
```

### Authorization request information

Extauth will by default transparently proxy the request to your authorization service.  It includes all of the headers, cookies, and the URL of the originally requested resource so that you can use that information to make an authorization decision.

### How to allow/deny a request

If you return `200 OK` from your authorization service, the request will be allowed to continue on through the other directives you have set up in your Caddyfile.  You can optionally add headers or cookies to the request that will then be propagated.

For example, you could check an authorization token, then update it with a new short-lived token by changing the `Authorization` header to the new token value.  This new token will replace the old in the original request.

A consequence of this behavior is that if you want cookies or headers to continue on with the request through your other directives and to their ultimate destination, you **must copy anything you want to continue on in the request chain into the response** that you return from your service.

If you return anything other than `200 OK`, the authorization will fail and a `401 Unauthorized` will be returned to the client immediately before evaluating any other Caddyfile directives.

### Advanced Syntax

You can optionally turn off headers, cookies, set a timeout, and skip TLS verification (for example, if you are using a self-signed cert):

```
extauth https://service {
  cookies false
  headers false
  timeout 30s
  insecure_skip_verify true  
}
```
