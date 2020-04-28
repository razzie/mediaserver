package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/razzie/mediaserver/siteinfo"
	"github.com/razzie/mediaserver/thumb"
)

// Response contains the content type and binary data response to a media request
type Response struct {
	ContentType string
	Data        []byte
	Error       string
}

func (r Response) serve(w http.ResponseWriter) {
	if len(r.Error) > 0 {
		http.Error(w, r.Error, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", r.ContentType)
	w.Header().Set("Content-Length", strconv.Itoa(len(r.Data)))
	w.Write(r.Data)
}

func respondSiteInfo(hostname string, src io.Reader) *Response {
	data, err := siteinfo.Get(src)
	if err != nil {
		return &Response{Error: err.Error()}
	}

	if len(data.ImageURL) > 0 {
		img, _ := url.Parse(data.ImageURL)
		data.ImageURL = fmt.Sprintf("http://%s/%s", hostname, img.Hostname()+img.RequestURI())
	}

	json, _ := json.MarshalIndent(data, "", "  ")
	return &Response{
		ContentType: "application/json",
		Data:        json,
	}
}

func respondThumbnail(src io.Reader) *Response {
	thumb, err := thumb.Get(src, "")
	if err != nil {
		return &Response{Error: err.Error()}
	}

	return &Response{
		ContentType: "image/jpeg",
		Data:        thumb,
	}
}
