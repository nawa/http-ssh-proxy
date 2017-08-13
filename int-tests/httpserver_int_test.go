package main_test

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nawa/http-ssh-proxy/config"
	"github.com/nawa/http-ssh-proxy/proxy"
	assert "github.com/stretchr/testify/require"
)

var handler http.Handler

func init() {
	cfg, _ := config.FromFile("./config.yml")
	handler = proxy.NewProxyServer(cfg)
}

func TestStartPage(t *testing.T) {
	recorder := doGet("/endpoint")
	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, fixture("server1.html"), recorder.Body.String())
}

func TestServer1(t *testing.T) {
	recorder := doGet("/server1/endpoint")
	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, fixture("server1.html"), recorder.Body.String())
}

func TestServer2(t *testing.T) {
	recorder := doGet("/server2/endpoint")
	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, fixture("server2.html"), recorder.Body.String())
}

func TestStartPageRedirectFullPath(t *testing.T) {
	recorder := doGet("/redirectFullPath")

	assert.Equal(t, http.StatusFound, recorder.Code)
	assert.Equal(t, `http://localhost:8888/server1/endpoint`, recorder.HeaderMap["Location"][0])
	assert.Equal(t, "<a href=\"http://localhost:8888/server1/endpoint\">Found</a>.\n\n", recorder.Body.String())
}

func TestServer1RedirectFullPath(t *testing.T) {
	recorder := doGet("/server1/redirectFullPath")

	assert.Equal(t, http.StatusFound, recorder.Code)
	assert.Equal(t, `http://localhost:8888/server1/endpoint`, recorder.HeaderMap["Location"][0])
	assert.Equal(t, "<a href=\"http://localhost:8888/server1/endpoint\">Found</a>.\n\n", recorder.Body.String())
}

func TestServer2RedirectFullPath(t *testing.T) {
	recorder := doGet("/server2/redirectFullPath")

	assert.Equal(t, http.StatusFound, recorder.Code)
	assert.Equal(t, `http://localhost:8888/server2/endpoint`, recorder.HeaderMap["Location"][0])
	assert.Equal(t, "<a href=\"http://localhost:8888/server2/endpoint\">Found</a>.\n\n", recorder.Body.String())
}

func TestStartPageRedirectRelativePath(t *testing.T) {
	recorder := doGet("/redirectRelativePath")

	assert.Equal(t, http.StatusFound, recorder.Code)
	assert.Equal(t, `http://localhost:8888/server1/endpoint`, recorder.HeaderMap["Location"][0])
	assert.Equal(t, "<a href=\"http://localhost:8888/server1/endpoint\">Found</a>.\n\n", recorder.Body.String())
}

func TestServer1RedirectRelativePath(t *testing.T) {
	recorder := doGet("/server1/redirectRelativePath")

	assert.Equal(t, http.StatusFound, recorder.Code)
	assert.Equal(t, `http://localhost:8888/server1/endpoint`, recorder.HeaderMap["Location"][0])
	assert.Equal(t, "<a href=\"http://localhost:8888/server1/endpoint\">Found</a>.\n\n", recorder.Body.String())
}

func TestServer2RedirectRelativePath(t *testing.T) {
	recorder := doGet("/server2/redirectRelativePath")

	assert.Equal(t, http.StatusFound, recorder.Code)
	assert.Equal(t, `http://localhost:8888/server2/endpoint`, recorder.HeaderMap["Location"][0])
	assert.Equal(t, "<a href=\"http://localhost:8888/server2/endpoint\">Found</a>.\n\n", recorder.Body.String())
}

func TestStartPageNotFound(t *testing.T) {
	recorder := doGet("/unknown")

	assert.Equal(t, http.StatusNotFound, recorder.Code)
	assert.Equal(t, "404 page not found\n", recorder.Body.String())
}

func TestServer1NotFound(t *testing.T) {
	recorder := doGet("/server1/unknown")

	assert.Equal(t, http.StatusNotFound, recorder.Code)
	assert.Equal(t, "404 page not found\n", recorder.Body.String())
}

func TestServer2NotFound(t *testing.T) {
	recorder := doGet("/server2/unknown")

	assert.Equal(t, http.StatusNotFound, recorder.Code)
	assert.Equal(t, "404 page not found\n", recorder.Body.String())
}

func TestServer3WithoutForwarding(t *testing.T) {
	recorder := doGet("/server3/endpoint")
	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, fixture("server3.html"), recorder.Body.String())
}

func doGet(path string) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", path, nil)
	request.Host = "localhost:8888"
	handler.ServeHTTP(recorder, request)
	return recorder
}

func fixture(name string) string {
	fixture, _ := ioutil.ReadFile("./fixtures/" + name)
	return string(fixture)
}

//TODO test gzip
// gzip
// request.Header.Add("Accept-Encoding", "gzip")
// request.Header.Add("Accept-Encoding", "deflate")
// handler.ServeHTTP(recorder, request)
// reader, _ := gzip.NewReader(recorder.Body)
// defer reader.Close()
// buf := new(bytes.Buffer)
// buf.ReadFrom(reader)
// response := buf.String()
