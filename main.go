package main

import (
	"flag"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func serveMedia(w http.ResponseWriter, r *http.Request) {
	if len(r.RequestURI) <= 1 {
		return
	}

	url := r.RequestURI[1:]
	req, err := http.NewRequest("GET", "http://"+url, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	req = req.WithContext(r.Context())
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if strings.HasPrefix(resp.Header.Get("Content-Type"), "text/html") {
		respondMetadata(r.Host, resp.Body).serve(w)
		return
	}

	respondThumbnail(resp.Body).serve(w)
}

func main() {
	port := flag.Int("port", 8080, "http port to listen on")
	flag.Parse()

	http.HandleFunc("/", serveMedia)
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(*port), nil))
}
