## EXTAUTH

[![Build Status](https://travis-ci.org/BTBurke/caddy-extauth.svg?branch=master)](https://travis-ci.org/BTBurke/caddy-extauth)

**This is a beta quality plugin for authorization that is undergoing testing.**

**External Authorization Middleware for Caddy**

This middleware implements an authorization layer for [Caddy](https://caddyserver.com) by using an external service you provide to handle authorization for each request.  Unlike other authorization middleware for Caddy like [JWT](https://github.com/BTBurke/caddy-jwt), this middleware provides you with more flexibility to implement your own authorization scheme.

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

Extauth will by proxy the request to your authorization service.  It includes all of the headers, cookies, and the host of the originally requested resource so that you can use that information to make an authorization decision.

### How to allow/deny a request

If you return `200 OK` from your authorization service, the request will be allowed to continue on through the other directives you have set up in your Caddyfile.  You can optionally add headers or cookies to the request that will then be propagated.

For example, you could check an authorization token, then update it with a new short-lived token by changing the `Authorization` header to the new token value.  This new token will replace the old in the original request.

A consequence of this behavior is that if you want cookies or headers to continue on with the request through your other directives and to their ultimate destination, you **must copy anything you want to continue on in the request chain into the response** that you return from your service.

If you return anything other than `200 OK`, the authorization will fail and a `401 Unauthorized` will be returned to the client immediately before evaluating any other Caddyfile directives.

## Simple vs. Router Mode

You can send a request to your authorization service in one of two modes: simple (default) or router mode.  In simple mode, the request always arrives at the root of your authorization API.  For example, if your user requests `https://example.com/deep/path?query=something` and your auth service is running on `http://localhost:9000`, you'll just receive a request to `http://localhost:9000/` without the path or query in the original request.  In order to recover the original path and query information, the simple mode sets a header `X-Auth-URL` with the full URL requested.  You can then parse that and take any action you like.

In router mode, all URL parameters are propagated to your auth service.  So in the above case, in router mode you would receive a request to your auth service like `http://localhost:9000/deep/path?query=something`.  This is useful if you want to parse the request using a router implementation in your language of choice to route the request to different authorization handlers.

To activate router mode, include the directive `router` in your config block:

```
extauth {
  proxy http://localhost:9000
  router
}
```

See the [examples](https://github.com/BTBurke/caddy-extauth/tree/master/examples) folder for how you might want to use each mode in your auth service.

### Advanced Syntax

You can optionally turn off headers, cookies, set a timeout, and skip TLS verification (for example, if you are using a self-signed cert):

```
extauth {
  proxy https://auth:9000
  cookies false
  headers false
  timeout 30s
  insecure_skip_verify true  
}
```
