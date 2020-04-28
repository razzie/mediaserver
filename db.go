package main

import (
	"encoding/json"
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
	data, err := db.client.Get(url).Result()
	if err != nil {
		return nil, err
	}

	var r media.Media
	err = json.Unmarshal([]byte(data), &r)
	if err != nil {
		return nil, err
	}

	return &r, nil
}

// SetMedia saves a Media
func (db *DB) SetMedia(url string, r *media.Media) error {
	data, err := json.Marshal(r)
	if err != nil {
		return err
	}

	return db.client.Set(url, string(data), db.ExpirationTime).Err()
}
