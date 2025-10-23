package main

import (
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
)

func proxyUrl(url string, removePrefixPath string) func(http.ResponseWriter, *http.Request) {
	// final url = url - remove
	return func(w http.ResponseWriter, r *http.Request) {
		client := &http.Client{}
		// Normalize removePrefixPath (remove trailing slash)
		removePrefixPath = strings.TrimSuffix(removePrefixPath, "/")

		// Trim the prefix from the request path
		urlPath := strings.TrimPrefix(r.URL.Path, removePrefixPath)
		if !strings.HasPrefix(urlPath, "/") && urlPath != "" {
			urlPath = "/" + urlPath
		}

		finalURL := url + urlPath
		if r.URL.RawQuery != "" {
			finalURL += "?" + r.URL.RawQuery
		}

		slog.Info("proxying request", "path", r.URL.Path, "trimmed", urlPath, "target", finalURL)

		req, err := http.NewRequest(r.Method, url+urlPath+"?"+r.URL.RawQuery, r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Copy headers
		for name, values := range r.Header {
			for _, value := range values {
				req.Header.Add(name, value)
			}
		}

		// Make the request
		resp, err := client.Do(req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		// Copy response headers
		for name, values := range resp.Header {
			for _, value := range values {
				w.Header().Add(name, value)
			}
		}

		w.WriteHeader(resp.StatusCode)
		_, err = io.Copy(w, resp.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	client := &http.Client{}
	req, err := http.NewRequest(r.Method, "https://api.themoviedb.org/3"+r.URL.Path+"?"+r.URL.RawQuery, r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Copy headers
	for name, values := range r.Header {
		for _, value := range values {
			req.Header.Add(name, value)
		}
	}

	// Make the request
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Copy response headers
	for name, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(name, value)
		}
	}

	w.WriteHeader(resp.StatusCode)
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func main() {
	port := "3000"
	// read port from environment variable if needed
	if envPort := os.Getenv("PORT"); envPort != "" {
		port = envPort
	}
	slog.Info("Starting proxy server", "port", port)
	http.HandleFunc("/3k-image/", proxyUrl("https://image.tmdb.org", "/3k-image"))
	http.HandleFunc("/live", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	http.HandleFunc("/", proxyUrl("https://api.themoviedb.org/3", ""))
	http.ListenAndServe(":"+port, nil)
}
