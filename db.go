package main

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/go-redis/redis/v7"
	"github.com/razzie/mediaserver/media"
)

// DB ...
type DB struct {
	ExpirationTime time.Duration
	client         *redis.Client
}

// NewDB returns a new DB
func NewDB(addr, password string, db int) (*DB, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	err := client.Ping().Err()
	if err != nil {
		client.Close()
		return nil, err
	}

	return &DB{
		ExpirationTime: time.Hour * 24,
		client:         client,
	}, nil
}

// GetMedia returns a saved Media
func (db *DB) GetMedia(url string) (*media.Media, error) {
	data, err := db.client.Get(urlToKey(url)).Result()
	if err != nil {
		return nil, err
	}

	var m media.Media
	err = json.Unmarshal([]byte(data), &m)
	if err != nil {
		return nil, err
	}

	return &m, nil
}

// SetMedia saves a Media
func (db *DB) SetMedia(url string, m *media.Media) error {
	data, err := json.Marshal(m)
	if err != nil {
		return err
	}

	expiration := db.ExpirationTime
	if m.Thumbnail == nil {
		expiration = time.Minute
	}

	return db.client.SetNX(urlToKey(url), string(data), expiration).Err()
}

func urlToKey(url string) string {
	url = strings.ToLower(url)
	if url[len(url)-1] == '/' {
		url = url[:len(url)-1]
	}
	return url
}
