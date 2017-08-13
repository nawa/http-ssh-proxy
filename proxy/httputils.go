package proxy

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strings"
)

var absLinkRegexp = regexp.MustCompile(`<[^>]+\b(href|src)[ \t\n]*=[ \t\n]*["'](/[^"']+)["'][^>]*>`)
var relativeLinkRegexp = regexp.MustCompile(`<[^>]+\b(href|src)[ \t\n]*=[ \t\n]*["']([^/][^"']+)["'][^>]*>`)

// Request : proxy http request that handles original
// request and returns response with replaced links
type Request struct {
	responseRecorder *httptest.ResponseRecorder
}

// Replacement : replacement pattern in text`
type Replacement struct {
	From string
	To   string
}

// ReplacementConfig : replacement configuration in response body
type ReplacementConfig struct {
	LinksBasePath             *url.URL
	ExpectedLocationHeader    string
	ExternalLinksReplacements []Replacement
}

// NewProxyRequest : proxy request constructor
func NewProxyRequest() *Request {
	return &Request{
		responseRecorder: httptest.NewRecorder(),
	}
}

// PerformRequest : performs proxy request and writes response with replacements
func (pr *Request) PerformRequest(requestHandler http.Handler,
	w http.ResponseWriter, request *http.Request, replaceConfig ReplacementConfig) error {
	requestHandler.ServeHTTP(pr.responseRecorder, request)

	return pr.writeToResponse(w, replaceConfig)
}

func (pr *Request) writeToResponse(responseWriter http.ResponseWriter, replaceConfig ReplacementConfig) error {
	for k, v := range pr.responseRecorder.HeaderMap {
		responseWriter.Header()[k] = v
	}
	isHTML := false
	for _, contentType := range pr.responseRecorder.HeaderMap["Content-Type"] {
		if strings.Contains(contentType, "html") {
			isHTML = true
			break
		}
	}
	replaceLocationHeader(pr, replaceConfig.ExpectedLocationHeader, replaceConfig.LinksBasePath)
	var writeError error
	if isHTML {
		writeError = writeHTMLBody(pr.responseRecorder, responseWriter, replaceConfig)
	} else {
		writeError = writeNonHTMLBody(responseWriter, pr)
	}

	if writeError != nil {
		return fmt.Errorf("Can't write to response: %v", writeError)
	}
	return nil
}

func writeNonHTMLBody(responseWriter http.ResponseWriter, pr *Request) (err error) {
	responseWriter.WriteHeader(pr.responseRecorder.Code)
	_, err = io.Copy(responseWriter, pr.responseRecorder.Body)
	return
}

func writeHTMLBody(responseRecorder *httptest.ResponseRecorder, responseWriter http.ResponseWriter, replaceConfig ReplacementConfig) (err error) {
	isGzip := false
	for _, contentEncoding := range responseRecorder.HeaderMap["Content-Encoding"] {
		if contentEncoding == "gzip" {
			isGzip = true
			break
		}
	}
	var sBody string
	if isGzip {
		reader, _ := gzip.NewReader(responseRecorder.Body)
		defer reader.Close() // nolint
		buf := new(bytes.Buffer)
		_, readErr := buf.ReadFrom(reader)
		if readErr != nil {
			return readErr
		}
		sBody = buf.String()
	} else {
		sBody = responseRecorder.Body.String()
	}
	sBody = replaceAbsoluteLinks(sBody, replaceConfig.LinksBasePath.String())
	sBody = replaceExternalLinks(sBody, replaceConfig.ExternalLinksReplacements)
	sBody = replaceRelativeLinks(sBody, replaceConfig.LinksBasePath.String())
	responseWriter.WriteHeader(responseRecorder.Code)
	if isGzip {
		gzipWriter := gzip.NewWriter(responseWriter)
		defer gzipWriter.Close() //nolint
		_, err = fmt.Fprint(gzipWriter, sBody)
	} else {
		_, err = fmt.Fprint(responseWriter, sBody)
	}
	return
}

func replaceLocationHeader(pr *Request, expectedLocationHeader string, linksBasePath *url.URL) {
	locationHeader := pr.responseRecorder.HeaderMap["Location"]
	for i, location := range locationHeader {
		if strings.HasPrefix(location, "/") {
			locationHeader[i] = linksBasePath.String() + location
		} else {
			locationHeader[i] = strings.Replace(location, expectedLocationHeader, linksBasePath.Host+"/"+linksBasePath.Path, 1)
		}
	}
}

func replaceAbsoluteLinks(html, basePath string) string {
	return addPrefixByRegexp(html, absLinkRegexp, 4, 5,
		func(link string) string {
			return basePath
		})
}

func replaceExternalLinks(html string, replacements []Replacement) (result string) {
	result = html
	for _, replacement := range replacements {
		linkRegexp := regexp.MustCompile(`<[^>]+\b(href|src)[ \t\n]*=[ \t\n]*["'].*(` + replacement.From + `).*["'][^>]*>`)
		result = replaceByRegexp(result, replacement.To, linkRegexp, 4, 5)
	}
	return
}

func replaceRelativeLinks(html, basePath string) string {
	return addPrefixByRegexp(html, relativeLinkRegexp, 4, 5,
		func(link string) (result string) {
			if !strings.HasPrefix(link, "http://") &&
				!strings.HasPrefix(link, "https://") {
				result = basePath + "/"
			}
			return
		})
}

func addPrefixByRegexp(s string, regexp *regexp.Regexp, groupStartIndex, groupEndIndex int,
	prefixBuilder func(link string) string) string {
	lastIndex := 0
	var buffer bytes.Buffer
	for _, v := range regexp.FindAllStringSubmatchIndex(s, -1) {
		originalValue := s[v[groupStartIndex]:v[groupEndIndex]]
		buffer.WriteString(s[lastIndex:v[groupStartIndex]])
		prefix := prefixBuilder(originalValue)
		if prefix != "" {
			buffer.WriteString(prefix)
		}
		buffer.WriteString(originalValue)
		lastIndex = v[groupEndIndex]
	}
	buffer.WriteString(s[lastIndex:])
	return buffer.String()
}

func replaceByRegexp(s, replaceTo string, regexp *regexp.Regexp, groupStartIndex, groupEndIndex int) string {
	lastIndex := 0
	var buffer bytes.Buffer
	for _, v := range regexp.FindAllStringSubmatchIndex(s, -1) {
		buffer.WriteString(s[lastIndex:v[groupStartIndex]])
		buffer.WriteString(replaceTo)
		lastIndex = v[groupEndIndex]
	}
	buffer.WriteString(s[lastIndex:])
	return buffer.String()
}
