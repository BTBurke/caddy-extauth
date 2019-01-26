package extauth

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/mholt/caddy"
	"github.com/mholt/caddy/caddyhttp/httpserver"
)

// Auth represents configuration information for the middleware
type Auth struct {
	Proxy              string
	Headers            bool
	Cookies            bool
	Timeout            time.Duration
	InsecureSkipVerify bool
	Router             bool
	Next               httpserver.Handler

	client *http.Client
}

func init() {
	caddy.RegisterPlugin("extauth", caddy.Plugin{
		ServerType: "http",
		Action:     Setup,
	})
}

// Setup is called by Caddy to parse the config block
func Setup(c *caddy.Controller) error {
	auth, err := parse(c)
	if err != nil {
		return err
	}

	c.OnStartup(func() error {
		fmt.Println("Extauth middleware is initiated")
		return nil
	})

	httpserver.GetConfig(c).AddMiddleware(func(next httpserver.Handler) httpserver.Handler {
		return &Auth{
			Proxy:              auth.Proxy,
			Headers:            auth.Headers,
			Cookies:            auth.Cookies,
			Timeout:            auth.Timeout,
			InsecureSkipVerify: auth.InsecureSkipVerify,
			Router:             auth.Router,
			Next:               next,
		}
	})

	return nil
}

func parse(c *caddy.Controller) (*Auth, error) {

	def := &Auth{
		Cookies:            true,
		Headers:            true,
		Timeout:            time.Duration(30 * time.Second),
		Router:             false,
		InsecureSkipVerify: false,
	}

	for c.Next() {
		args := c.RemainingArgs()
		switch len(args) {
		case 0:
			// no argument passed, check the config block
			var err error
			for c.NextBlock() {
				switch c.Val() {
				case "router":
					def.Router = true
				case "proxy":
					if !c.NextArg() {
						// we are expecting a value
						return nil, c.ArgErr()
					}
					def.Proxy = c.Val()
					if c.NextArg() {
						// we are expecting only one value.
						return nil, c.ArgErr()
					}
				case "cookies":
					if !c.NextArg() {
						return nil, c.ArgErr()
					}
					def.Cookies, err = strconv.ParseBool(c.Val())
					if err != nil {
						return nil, c.ArgErr()
					}
					if c.NextArg() {
						return nil, c.ArgErr()
					}
				case "headers":
					if !c.NextArg() {
						return nil, c.ArgErr()
					}
					def.Headers, err = strconv.ParseBool(c.Val())
					if err != nil {
						return nil, c.ArgErr()
					}
					if c.NextArg() {
						return nil, c.ArgErr()
					}
				case "insecure_skip_verify":
					def.InsecureSkipVerify = true
				case "timeout":
					if !c.NextArg() {
						return nil, c.ArgErr()
					}
					def.Timeout, err = time.ParseDuration(c.Val())
					if err != nil {
						return nil, c.ArgErr()
					}
				default:
					return nil, c.Errf("unsupported directive: '%s'", args[0])
				}
			}
		case 1:
			log.Printf("got proxy: %s", args[0])
			def.Proxy = args[0]
		default:
			// we want only one argument max
			log.Println("got an error for no args")
			return nil, c.ArgErr()
		}
	}

	if len(def.Proxy) == 0 {
		return nil, fmt.Errorf("No proxy for extauth configured")
	}

	return def, nil
}
