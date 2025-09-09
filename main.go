package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
)

func handler(w http.ResponseWriter, r *http.Request) {
	log.Println("Received request:", r.URL)

	// Parse query params
	q := r.URL.Query()
	forwardURL := q.Get("forward_url")
	apiKey := q.Get("api_key")

	if forwardURL == "" || apiKey == "" {
		http.Error(w, "missing forward_url or api_key", http.StatusBadRequest)
		return
	}

	// Remove control params so we can forward the rest
	q.Del("forward_url")
	q.Del("api_key")

	// Parse and validate forward_url
	target, err := url.Parse(forwardURL)
	if err != nil || !(target.Scheme == "http" || target.Scheme == "https") {
		http.Error(w, "invalid forward_url", http.StatusBadRequest)
		return
	}

	// Append remaining query parameters to the target URL
	target.RawQuery = q.Encode()

	// Prepare log of remaining query params
	var params []string
	for k, vs := range q {
		params = append(params, fmt.Sprintf("%s=%s", k, strings.Join(vs, ",")))
	}

	log.Printf("Forwarding request to %s with query params: %s", forwardURL, strings.Join(params, "&"))

	// Forward the request
	req, err := http.NewRequest(http.MethodGet, target.String(), nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Set headers
	req.Header = r.Header.Clone()
	req.Header.Set("X-Flyt-API-Key", apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// Copy response headers
	for k, vv := range resp.Header {
		for _, v := range vv {
			w.Header().Add(k, v)
		}
	}

	// Copy response status & body
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func main() {
	http.HandleFunc("/", handler)
	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
