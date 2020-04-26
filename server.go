package main

import (
	"log"
	"net"
	"net/http"
	"strings"
)

// Server ...
type Server struct {
	db  *DB
	mux http.ServeMux
}

// NewServer returns a new server
func NewServer(db *DB) *Server {
	srv := &Server{db: db}
	srv.mux.HandleFunc("/", srv.handleRequest)
	srv.mux.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Not found", http.StatusNotFound)
	})
	return srv
}

func (srv *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	srv.mux.ServeHTTP(w, r)
}

func (srv *Server) handleRequest(w http.ResponseWriter, r *http.Request) {
	logRequest(r)

	if len(r.RequestURI) <= 1 {
		return
	}

	url := r.RequestURI[1:]
	if strings.HasPrefix(url, "http:/") {
		http.Redirect(w, r, "/"+url[7:], http.StatusSeeOther)
		return
	} else if strings.HasPrefix(url, "https:/") {
		http.Redirect(w, r, "/"+url[8:], http.StatusSeeOther)
		return
	}

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

func logRequest(r *http.Request) {
	ip := r.Header.Get("X-REAL-IP")
	if len(ip) == 0 {
		ip, _, _ = net.SplitHostPort(r.RemoteAddr)
	}

	log.Println(ip, r.RequestURI)
}
