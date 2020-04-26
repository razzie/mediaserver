package main

import (
	"net/http"
	"strings"
)

// Server ...
type Server struct {
	db *DB
}

// NewServer returns a new server
func NewServer(db *DB) *Server {
	return &Server{db: db}
}

func (srv *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if len(r.RequestURI) <= 1 {
		return
	}

	url := r.RequestURI[1:]

	cached, _ := srv.db.GetResponse(url)
	if cached != nil {
		cached.serve(w)
		return
	}

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
		cached = respondMetadata(r.Host, resp.Body)
	} else {
		cached = respondThumbnail(resp.Body)
	}

	cached.serve(w)
	srv.db.SetResponse(url, cached)
}
