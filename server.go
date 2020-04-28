package main

import (
	"log"
	"net"
	"net/http"
	"strings"

	"github.com/razzie/mediaserver/media"
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

	url, changed := removeSchemeFromURL(r.RequestURI[1:])
	if changed {
		http.Redirect(w, r, "/"+url, http.StatusSeeOther)
		return
	}

	cached, _ := srv.db.GetMedia(url)
	if cached != nil {
		cached.ServeHTTP(w, r)
		return
	}

	resp, err := media.GetFromURL(r.Context(), url)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	} else {
		resp.ServeHTTP(w, r)
	}

	if resp != nil {
		srv.db.SetMedia(url, resp)
	}
}

func logRequest(r *http.Request) {
	ip := r.Header.Get("X-REAL-IP")
	if len(ip) == 0 {
		ip, _, _ = net.SplitHostPort(r.RemoteAddr)
	}

	log.Println(ip, r.RequestURI)
}

func removeSchemeFromURL(url string) (result string, changed bool) {
	if index := strings.Index(url, ":/"); index != -1 {
		if url[index+1:index+2] == "/" {
			return url[index+2:], true
		}
		return url[index+1:], true
	}
	return url, false
}
