package main

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type StorageClient struct {
	client *minio.Client
	cfg    *Config
}

func NewStorageClient(cfg *Config) (*StorageClient, error) {
	u, err := url.Parse(cfg.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid RUSTFS_BASE_URL: %w", err)
	}

	endpoint := u.Host
	useSSL := strings.EqualFold(u.Scheme, "https")

	mc, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create storage client: %w", err)
	}

	return &StorageClient{client: mc, cfg: cfg}, nil
}

var extContentTypes = map[string]string{
	".txt":     "text/plain; charset=utf-8",
	".md":      "text/markdown; charset=utf-8",
	".csv":     "text/csv; charset=utf-8",
	".json":    "application/json",
	".yaml":    "text/yaml; charset=utf-8",
	".yml":     "text/yaml; charset=utf-8",
	".toml":    "text/plain; charset=utf-8",
	".xml":     "application/xml",
	".html":    "text/html; charset=utf-8",
	".htm":     "text/html; charset=utf-8",
	".js":      "application/javascript",
	".ts":      "text/typescript; charset=utf-8",
	".go":      "text/plain; charset=utf-8",
	".py":      "text/plain; charset=utf-8",
	".rs":      "text/plain; charset=utf-8",
	".java":    "text/plain; charset=utf-8",
	".c":       "text/plain; charset=utf-8",
	".cpp":     "text/plain; charset=utf-8",
	".h":       "text/plain; charset=utf-8",
	".sh":      "text/plain; charset=utf-8",
	".bash":    "text/plain; charset=utf-8",
	".fish":    "text/plain; charset=utf-8",
	".zsh":     "text/plain; charset=utf-8",
	".sql":     "text/plain; charset=utf-8",
	".graphql": "text/plain; charset=utf-8",
	".tf":      "text/plain; charset=utf-8",
	".log":     "text/plain; charset=utf-8",
	".env":     "text/plain; charset=utf-8",
}

func contentTypeForKey(key string) string {
	ext := strings.ToLower(filepath.Ext(key))
	if ct, ok := extContentTypes[ext]; ok {
		return ct
	}
	return "text/plain; charset=utf-8"
}

func (s *StorageClient) PutObject(ctx context.Context, bucket, key, content string) error {
	r := strings.NewReader(content)
	_, err := s.client.PutObject(ctx, bucket, key, r, int64(len(content)), minio.PutObjectOptions{
		ContentType: contentTypeForKey(key),
	})
	if err != nil {
		return fmt.Errorf("put %s/%s: %w", bucket, key, err)
	}
	return nil
}

func (s *StorageClient) GetObject(ctx context.Context, bucket, key string) ([]byte, error) {
	obj, err := s.client.GetObject(ctx, bucket, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("get %s/%s: %w", bucket, key, err)
	}
	defer obj.Close()

	data, err := io.ReadAll(obj)
	if err != nil {
		return nil, fmt.Errorf("read %s/%s: %w", bucket, key, err)
	}

	return data, nil
}
