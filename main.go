package main

import (
	"flag"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/razzie/mediaserver/thumb"
)

// Command-line args
var (
	RedisConnStr  string
	Port          int
	CacheDuration time.Duration
)

func main() {
	flag.StringVar(&RedisConnStr, "redis", "redis://localhost:6379", "Redis connection string")
	flag.IntVar(&Port, "port", 8080, "HTTP port to listen on")
	flag.IntVar(&thumb.Quality, "thumb-quality", 90, "Quality of the thumbnail images (1-100)")
	flag.UintVar(&thumb.Size, "thumb-size", 256, "Maximum width or height of thumbnail images")
	flag.DurationVar(&CacheDuration, "cache-duration", time.Hour*24, "Thumbnail cache expiration time")
	flag.Parse()

	db, err := NewDB(RedisConnStr)
	if err != nil {
		log.Fatalln("failed to connect to database:", err)
	}

	db.ExpirationTime = CacheDuration

	server := NewServer(db)
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(Port), server))
}
