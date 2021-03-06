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
	RedisAddr     string
	RedisPw       string
	RedisDb       int
	Port          int
	CacheDuration time.Duration
)

func main() {
	flag.StringVar(&RedisAddr, "redis-addr", "localhost:6379", "Redis hostname:port")
	flag.StringVar(&RedisPw, "redis-pw", "", "Redis password")
	flag.IntVar(&RedisDb, "redis-db", 0, "Redis database (0-15)")
	flag.IntVar(&Port, "port", 8080, "HTTP port to listen on")
	flag.IntVar(&thumb.Quality, "thumb-quality", 90, "Quality of the thumbnail images (1-100)")
	flag.UintVar(&thumb.Size, "thumb-size", 256, "Maximum width or height of thumbnail images")
	flag.DurationVar(&CacheDuration, "cache-duration", time.Hour*24, "Thumbnail cache expiration time")
	flag.Parse()

	db, err := NewDB(RedisAddr, RedisPw, RedisDb)
	if err != nil {
		log.Fatalln("failed to connect to database:", err)
	}

	db.ExpirationTime = CacheDuration

	server := NewServer(db)
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(Port), server))
}
