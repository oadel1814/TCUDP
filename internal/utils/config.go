package utils

import "time"

type Config struct {
	Timeout    time.Duration
	MaxRetries int
	ListenAddr string
}

func DefaultConfig() *Config {
	return &Config{
		Timeout:    1000 * time.Millisecond,
		MaxRetries: 5,
		ListenAddr: "127.0.0.1:8080",
	}
}
