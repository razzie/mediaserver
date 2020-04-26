package main

import (
	"flag"
	"log"
	"net/http"
	"strconv"
)

func main() {
	port := flag.Int("port", 8080, "http port to listen on")
	flag.Parse()

	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(*port), &Server{}))
}
