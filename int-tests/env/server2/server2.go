package main

import (
	"io"
	"net/http"
	"net/url"
	"os"
)

func main() {
	http.HandleFunc("/endpoint", handler)
	http.HandleFunc("/redirectFullPath", redirectFullPathHandler)
	http.HandleFunc("/redirectRelativePath", redirectRelativePathHandler)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	f, err := os.Open("server2/server2.html")
	if err != nil {
		panic(err)
	}
	defer f.Close() // nolint
	io.Copy(w, f)   // nolint
}

func redirectFullPathHandler(w http.ResponseWriter, r *http.Request) {
	var scheme string
	if r.TLS != nil {
		scheme = "https"
	} else {
		scheme = "http"
	}

	redirectPath := url.URL{
		Scheme: scheme,
		Host:   r.Host,
		Path:   "/endpoint",
	}
	http.Redirect(w, r, redirectPath.String(), http.StatusFound)
}

func redirectRelativePathHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/endpoint", http.StatusFound)
}
