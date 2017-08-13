package proxy

import (
	"net/url"
	"testing"

	"github.com/nawa/http-ssh-proxy/config"
	assert "github.com/stretchr/testify/require"
)

func TestParseHostName(t *testing.T) {
	cfg, _ := config.FromFile("../config/testdata/config.yml")
	urlValue, _ := url.Parse("http://localhost:8080")
	hostName, tail := parseHostName(urlValue, cfg)
	assert.Equal(t, "master", hostName)
	assert.Equal(t, "", tail)

	urlValue, _ = url.Parse("http://localhost:8080/")
	hostName, tail = parseHostName(urlValue, cfg)
	assert.Equal(t, "master", hostName)
	assert.Equal(t, "", tail)

	urlValue, _ = url.Parse("http://localhost:8080/favicon.ico")
	hostName, tail = parseHostName(urlValue, cfg)
	assert.Equal(t, "master", hostName)
	assert.Equal(t, "favicon.ico", tail)

	urlValue, _ = url.Parse("http://localhost:8080/unknown/path/to/something")
	hostName, tail = parseHostName(urlValue, cfg)
	assert.Equal(t, "master", hostName)
	assert.Equal(t, "unknown/path/to/something", tail)

	urlValue, _ = url.Parse("http://localhost:8080/worker-2")
	hostName, tail = parseHostName(urlValue, cfg)
	assert.Equal(t, "worker-2", hostName)
	assert.Equal(t, "", tail)

	urlValue, _ = url.Parse("http://localhost:8080/worker-2/")
	hostName, tail = parseHostName(urlValue, cfg)
	assert.Equal(t, "worker-2", hostName)
	assert.Equal(t, "", tail)

	urlValue, _ = url.Parse("http://localhost:8080/worker-2/path/to/something")
	hostName, tail = parseHostName(urlValue, cfg)
	assert.Equal(t, "worker-2", hostName)
	assert.Equal(t, "path/to/something", tail)
}
