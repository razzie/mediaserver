package main

import (
	"encoding/json"
	"flag"
	"io"
	"log"
	"mime"
	"net/http"
	"strconv"
	"strings"

	browser "github.com/EDDYCJY/fake-useragent"
	"github.com/razzie/mediaserver/og"
	"github.com/razzie/mediaserver/thumb"
)

// https://gist.github.com/rjz/fe283b02cbaa50c5991e1ba921adf7c9
func hasContentType(r *http.Response, mimetype string) bool {
	contentType := r.Header.Get("Content-type")
	if contentType == "" {
		return mimetype == "application/octet-stream"
	}

	for _, v := range strings.Split(contentType, ",") {
		t, _, err := mime.ParseMediaType(v)
		if err != nil {
			break
		}
		if t == mimetype {
			return true
		}
	}
	return false
}

func serveMedia(w http.ResponseWriter, r *http.Request) {
	if len(r.URL.Path) <= 1 {
		return
	}

	url := r.URL.Path[1:]
	url = strings.Replace(url, ":/", "://", 1)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	req.Header.Set("User-Agent", browser.Random())

	resp, err := http.DefaultClient.Do(req.WithContext(r.Context()))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if hasContentType(resp, "text/html") {
		serveMetadata(w, r.Host, resp.Body)
		return
	}

	serveThumbnail(w, resp.Body)
}

func serveMetadata(w http.ResponseWriter, hostname string, src io.Reader) {
	data, err := og.Get(src)
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
