package ssh

import (
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"golang.org/x/crypto/ssh"
)

// Tunnel : ssh tunnel
type Tunnel struct {
	Server string
	Remote string

	SSHClientConfig *ssh.ClientConfig
}

// NewTunnelByUserPassword : tunnel constructor using user/password
func NewTunnelByUserPassword(server, remote, user, password string) *Tunnel {
	return &Tunnel{
		Server: server,
		Remote: remote,
		SSHClientConfig: &ssh.ClientConfig{
			User:            user,
			Auth:            []ssh.AuthMethod{ssh.Password(password)},
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		},
	}
}

// NewTunnelByUserKey : tunnel constructor using user/key
func NewTunnelByUserKey(server, remote, user, key string) (*Tunnel, error) {
	authByCertificate, err := publicKeyFile(key)
	if err != nil {
		return nil, fmt.Errorf("Can't import private key: %v", err)
	}
	tunnel := &Tunnel{
		Server: server,
		Remote: remote,
		SSHClientConfig: &ssh.ClientConfig{
			User:            user,
			Auth:            []ssh.AuthMethod{authByCertificate},
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		},
	}
	return tunnel, nil
}

// CreateReverseProxy : creates http reverse proxy that serves your http requests through configured ssh connection
func (tunnel *Tunnel) CreateReverseProxy() (*httputil.ReverseProxy, error) {
	serverConn, err := ssh.Dial("tcp", tunnel.Server, tunnel.SSHClientConfig)
	if err != nil {
		return nil, fmt.Errorf("Server dial error: %v", err)
	}

	reverseProxy := httputil.NewSingleHostReverseProxy(&url.URL{
		Scheme: "http",
		Host:   tunnel.Remote,
	})

	reverseProxy.Transport = &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			remoteConn, err := serverConn.Dial(network, addr)
			if err != nil {
				return nil, fmt.Errorf("Remote dial error: %v", err)
			}
			return remoteConn, err
		},
	}
	return reverseProxy, nil
}

func publicKeyFile(file string) (ssh.AuthMethod, error) {
	buffer, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	key, err := ssh.ParsePrivateKey(buffer)
	if err != nil {
		return nil, err
	}
	return ssh.PublicKeys(key), nil
}
