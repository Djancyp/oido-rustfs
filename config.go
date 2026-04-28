package main

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	BaseURL       string
	AccessKey     string
	SecretKey     string
	Buckets       []string
	DefaultBucket string
}

func LoadConfig() (*Config, error) {
	baseURL := os.Getenv("OIDO_RUSTFS_BASE_URL")
	if baseURL == "" {
		return nil, fmt.Errorf("OIDO_RUSTFS_BASE_URL is required")
	}

	accessKey := os.Getenv("OIDO_RUSTFS_ACCESS_KEY")
	if accessKey == "" {
		return nil, fmt.Errorf("OIDO_RUSTFS_ACCESS_KEY is required")
	}

	secretKey := os.Getenv("OIDO_RUSTFS_SECRET_KEY")
	if secretKey == "" {
		return nil, fmt.Errorf("OIDO_RUSTFS_SECRET_KEY is required")
	}

	bucketEnv := os.Getenv("OIDO_RUSTFS_BUCKET")
	if bucketEnv == "" {
		bucketEnv = "chat-attachments"
	}

	parts := strings.Split(bucketEnv, ",")
	buckets := make([]string, 0, len(parts))
	for _, b := range parts {
		if b = strings.TrimSpace(b); b != "" {
			buckets = append(buckets, b)
		}
	}

	return &Config{
		BaseURL:       baseURL,
		AccessKey:     accessKey,
		SecretKey:     secretKey,
		Buckets:       buckets,
		DefaultBucket: buckets[0],
	}, nil
}

func (c *Config) IsAllowedBucket(bucket string) bool {
	for _, b := range c.Buckets {
		if b == bucket {
			return true
		}
	}
	return false
}
