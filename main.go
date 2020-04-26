package main

import (
	"encoding/json"
	"flag"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/razzie/mediaserver/og"
	"github.com/razzie/mediaserver/thumb"
)

func serveMedia(w http.ResponseWriter, r *http.Request) {
	if len(r.RequestURI) <= 1 {
		return
	}

	url := r.RequestURI[1:]
	url = strings.Replace(url, ":/", "://", 1)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	//req.Header.Set("User-Agent", browser.Random())
	//req.Header.Set("accept", "text/html,image/*")

	resp, err := http.DefaultClient.Do(req /*.WithContext(r.Context())*/)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if strings.HasPrefix(resp.Header.Get("Content-Type"), "text/html") {
		serveMetadata(w, r.Host, url, resp.Body)
		return
	}

	serveThumbnail(w, resp.Body)
}

func serveMetadata(w http.ResponseWriter, hostname, url string, src io.Reader) {
	data, err := og.Get(url, src)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(data.ImageURL) > 0 {
		data.ImageURL = hostname + "/" + data.ImageURL
	}

	json, _ := json.MarshalIndent(data, "", "  ")
	w.Header().Set("Content-Type", "application/json")
	w.Write(json)
}

func serveThumbnail(w http.ResponseWriter, src io.Reader) {
	thumb, err := thumb.Get(src)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("Content-Length", strconv.Itoa(len(thumb)))
	w.Write(thumb)
}

func main() {
	port := flag.Int("port", 8080, "http port to listen on")
	flag.Parse()

	http.HandleFunc("/", serveMedia)
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(*port), nil))
}
