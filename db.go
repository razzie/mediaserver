package main

import (
	"encoding/json"
	"time"

	"github.com/go-redis/redis/v7"
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

// GetResponse returns a saved Response
func (db *DB) GetResponse(url string) (*Response, error) {
	data, err := db.client.Get(url).Result()
	if err != nil {
		return nil, err
	}

	var r Response
	err = json.Unmarshal([]byte(data), &r)
	if err != nil {
		return nil, err
	}

	return &r, nil
}

// SetResponse saves a Response
func (db *DB) SetResponse(url string, r *Response) error {
	data, err := json.Marshal(r)
	if err != nil {
		return err
	}

	return db.client.Set(url, string(data), db.ExpirationTime).Err()
}
