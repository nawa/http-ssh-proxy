package proxy

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"runtime/debug"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/justinas/alice"
	"github.com/nawa/http-ssh-proxy/config"
	"github.com/nawa/http-ssh-proxy/ssh"
)

// HTTPServer : proxy http server - central point of application
type HTTPServer struct {
	Config      *config.Config
	rootHandler http.Handler
}

// HTTPError : http error with message and code
type HTTPError struct {
	Code    int
	Message string
}

// NewProxyServer : proxy http server constructor
func NewProxyServer(config *config.Config) *HTTPServer {
	rootHandler := alice.New(recoverHandler).
		Then(proxyHandler(config))
	return &HTTPServer{Config: config, rootHandler: rootHandler}
}

// Start : starts proxy http server
func (httpServer *HTTPServer) Start() {
	http.Handle("/", httpServer.rootHandler)
	err := http.ListenAndServe(fmt.Sprintf("localhost:%v", httpServer.Config.AppPort), nil)
	if err != nil {
		log.Fatal(err)
	}
}

func (httpServer *HTTPServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	httpServer.rootHandler.ServeHTTP(w, r)
}

func recoverHandler(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				httpError, ok := err.(HTTPError)
				if ok {
					log.Warnf("An error handled %s, %v", httpError.Message, httpError.Code)
					http.Error(w, httpError.Message, httpError.Code)
				} else {
					log.Errorf("An unknown error was handled for [%s] %s: %v\n Stack trace:\n%s",
						r.Method, r.RequestURI, err, debug.Stack())
					http.Error(w,
						fmt.Sprintf("%s: %v", http.StatusText(http.StatusInternalServerError), err),
						http.StatusInternalServerError)
				}
			}
		}()
		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}

func proxyHandler(config *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		hostName, tail := parseHostName(r.URL, config)
		host, ok := config.Hosts[hostName]
		if !ok {
			panic(HTTPError{Message: fmt.Sprintf("Host not found for '%s'", hostName), Code: http.StatusNotFound})
		} else {
			var reverseProxy = createReverseProxy(host)
			var originalHost = r.Host
			var remoteHost = host.Address
			var rw = NewProxyRequest()

			if len(tail) > 0 {
				r.URL.Path = tail
				r.RequestURI = r.URL.RequestURI()
			}
			r.Host = remoteHost

			proxyBasePath := &url.URL{
				Scheme: "http",
				Host:   originalHost,
				Path:   hostName,
			}

			var linksReplacements []Replacement
			for hostName, desc := range config.Hosts {
				linksReplacements = append(linksReplacements, Replacement{
					From: desc.Address,
					To:   originalHost + "/" + hostName,
				})
			}

			replaceConfig := ReplacementConfig{
				LinksBasePath:             proxyBasePath,
				ExpectedLocationHeader:    remoteHost,
				ExternalLinksReplacements: linksReplacements,
			}

			err := rw.PerformRequest(reverseProxy, w, r, replaceConfig)
			if err != nil {
				log.Errorf("Request to %s has been failed. Error: %v", host.Address, err)
				panic(HTTPError{Message: err.Error(), Code: http.StatusInternalServerError})
			}

			if host.Forwarding != nil {
				log.Infof("Request to %s was successfully proxied using ssh through %s", host.Address, host.Forwarding.Server)
			} else {
				log.Infof("Request to %s was successfully performed", host.Address)
			}

		}
	}
}

func createReverseProxy(host config.Host) (reverseProxy *httputil.ReverseProxy) {
	if host.Forwarding != nil {
		tunnel, err := createSSHTunnelFromConfig(host)
		if err != nil {
			log.Panicf("Can't create ssh tunnel for forwarding : %v", err)
		}
		reverseProxy, err = tunnel.CreateReverseProxy()
		if err != nil {
			log.Panicf("Can't forward request: %v", err)
		}
	} else {
		reverseProxy = httputil.NewSingleHostReverseProxy(&url.URL{
			Scheme: "http",
			Host:   host.Address,
		})
	}
	return reverseProxy
}

func createSSHTunnelFromConfig(configHost config.Host) (*ssh.Tunnel, error) {
	if configHost.Forwarding.PrivateKey != nil {
		tunnel, err := ssh.NewTunnelByUserKey(configHost.Forwarding.Server, configHost.Address,
			configHost.Forwarding.User, *configHost.Forwarding.PrivateKey)
		if err != nil {
			return nil, err
		}
		return tunnel, nil
	} else if configHost.Forwarding.Password != nil {
		return ssh.NewTunnelByUserPassword(configHost.Forwarding.Server, configHost.Address,
			configHost.Forwarding.User, *configHost.Forwarding.Password), nil
	}
	return nil, fmt.Errorf("Unknown forwarding type")
}

func parseHostName(url *url.URL, config *config.Config) (hostName, tail string) {
	tail = ""
	if len(url.Path) == 0 ||
		(len(url.Path) == 1 && url.Path[0] == '/') {
		hostName = config.StartPage
	} else {
		withoutFirstSlash := url.Path[1:]
		parts := strings.Split(withoutFirstSlash, "/")
		hostName = parts[0]
		if _, ok := config.Hosts[hostName]; !ok {
			hostName = config.StartPage
			tail = strings.Join(parts, "/")
		} else {
			tail = strings.Join(parts[1:], "/")
		}
		log.Printf("'%v' host name parsed from URI", hostName)
	}
	return
}
