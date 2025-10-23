package main

import (
	"io"
	"log/slog"
	"net/http"
	"os"
)

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
	http.HandleFunc("/", handler)
	http.ListenAndServe(":"+port, nil)
}
