package main

import (
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
)

func handler(w http.ResponseWriter, r *http.Request) {
	log.Println("Received request:", r.URL)

	// Extract control params from query string
	q := r.URL.Query()
	forwardURL := q.Get("forward_url")
	apiKey := q.Get("api_key")

	if forwardURL == "" || apiKey == "" {
		http.Error(w, "missing forward_url or api_key", http.StatusBadRequest)
		return
	}

	// Parse form values from request body (POST form)
	if err := r.ParseForm(); err != nil {
		http.Error(w, "failed to parse form: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Remove control params from form in case they exist there
	r.Form.Del("forward_url")
	r.Form.Del("api_key")

	// Encode form values for POST
	formData := r.Form.Encode()

	// Parse and validate forward_url
	target, err := url.Parse(forwardURL)
	if err != nil || !(target.Scheme == "http" || target.Scheme == "https") {
		http.Error(w, "invalid forward_url", http.StatusBadRequest)
		return
	}

	log.Printf("Forwarding POST request to %s with form data: %s", forwardURL, formData)

	// Create POST request
	req, err := http.NewRequest(http.MethodPost, target.String(), strings.NewReader(formData))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Copy original headers, add API key, and set content type
	req.Header = r.Header.Clone()
	req.Header.Set("X-Flyt-API-Key", apiKey)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Send the request
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
