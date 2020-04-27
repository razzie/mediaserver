package main

import (
	"flag"
	"log"
	"net/http"
	"strconv"

	"github.com/razzie/mediaserver/thumb"
)

func main() {
	redisAddr := flag.String("redis-addr", "localhost:6379", "Redis hostname:port")
	redisPw := flag.String("redis-pw", "", "Redis password")
	redisDb := flag.Int("redis-db", 0, "Redis database (0-15)")
	port := flag.Int("port", 8080, "http port to listen on")
	flag.IntVar(&thumb.Quality, "-thumb-quality", 10, "Quality of the thumbnail images (1-100)")
	flag.UintVar(&thumb.Size, "-thumb-size", 256, "Maximum width or height of thumbnail images")
	flag.Parse()

	db, err := NewDB(*redisAddr, *redisPw, *redisDb)
	if err != nil {
		log.Fatalln("failed to connect to database:", err)
	}

	server := NewServer(db)
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(*port), server))
}
