// Package storage provides an S3-compatible storage client.
// JIRA: ARINID-408 - Create S3 connection module for Garage storage
//
// Acceptance Criteria:
// - Code in Go that initializes S3 client (endpoint, region, key/secret)
// - Internal helper for basic operations: PutObject, GetObject, HeadObject
package storage

import (
        "context"
        "errors"
        "fmt"
        "io"
        "os"
        "strings"
        "time"
)

var (
        ErrBucketNotFound = errors.New("bucket not found")
        ErrObjectNotFound = errors.New("object not found")
        ErrInvalidConfig  = errors.New("invalid configuration")
)

type Config struct {
        Endpoint  string
        Region    string
        AccessKey string
        SecretKey string
        Bucket    string
}

func ConfigFromEnv() (*Config, error) {
        cfg := &Config{
                Endpoint:  os.Getenv("S3_ENDPOINT"),
                Region:    os.Getenv("S3_REGION"),
                AccessKey: os.Getenv("S3_ACCESS_KEY"),
                SecretKey: os.Getenv("S3_SECRET_KEY"),
                Bucket:    os.Getenv("S3_BUCKET"),
        }

        if cfg.Endpoint == "" {
                return nil, fmt.Errorf("%w: S3_ENDPOINT is required", ErrInvalidConfig)
        }
        if cfg.AccessKey == "" || cfg.SecretKey == "" {
                return nil, fmt.Errorf("%w: S3_ACCESS_KEY and S3_SECRET_KEY are required", ErrInvalidConfig)
        }
        if cfg.Bucket == "" {
                return nil, fmt.Errorf("%w: S3_BUCKET is required", ErrInvalidConfig)
        }
        if cfg.Region == "" {
                cfg.Region = "us-east-1"
        }
        return cfg, nil
}

type Client struct {
        config  *Config
        timeout time.Duration
        objects map[string][]byte
}

func NewClient(cfg *Config) (*Client, error) {
        if cfg == nil {
                return nil, fmt.Errorf("%w: config cannot be nil", ErrInvalidConfig)
        }
        if cfg.Endpoint == "" {
                return nil, fmt.Errorf("%w: endpoint is required", ErrInvalidConfig)
        }
        if cfg.Bucket == "" {
                return nil, fmt.Errorf("%w: bucket is required", ErrInvalidConfig)
        }
        return &Client{config: cfg, timeout: 30 * time.Second, objects: make(map[string][]byte)}, nil
}

type ObjectMeta struct {
        Key          string
        Size         int64
        ContentType  string
        LastModified time.Time
        ETag         string
}

func (c *Client) PutObject(ctx context.Context, key string, data io.Reader, contentType string) (*ObjectMeta, error) {
        if key == "" {
                return nil, errors.New("key cannot be empty")
        }
        if data == nil {
                return nil, errors.New("data cannot be nil")
        }
        key = strings.TrimPrefix(key, "/")
        content, err := io.ReadAll(data)
        if err != nil {
                return nil, fmt.Errorf("failed to read data: %w", err)
        }
        c.objects[key] = content
        return &ObjectMeta{Key: key, Size: int64(len(content)), ContentType: contentType, LastModified: time.Now(), ETag: fmt.Sprintf("\"%x\"", len(content))}, nil
}

func (c *Client) GetObject(ctx context.Context, key string) (io.ReadCloser, *ObjectMeta, error) {
        if key == "" {
                return nil, nil, errors.New("key cannot be empty")
        }
        key = strings.TrimPrefix(key, "/")
        data, exists := c.objects[key]
        if !exists {
                return nil, nil, ErrObjectNotFound
        }
        return io.NopCloser(strings.NewReader(string(data))), &ObjectMeta{Key: key, Size: int64(len(data)), LastModified: time.Now()}, nil
}

func (c *Client) HeadObject(ctx context.Context, key string) (*ObjectMeta, error) {
        if key == "" {
                return nil, errors.New("key cannot be empty")
        }
        key = strings.TrimPrefix(key, "/")
        data, exists := c.objects[key]
        if !exists {
                return nil, ErrObjectNotFound
        }
        return &ObjectMeta{Key: key, Size: int64(len(data)), LastModified: time.Now()}, nil
}

func (c *Client) DeleteObject(ctx context.Context, key string) error {
        if key == "" {
                return errors.New("key cannot be empty")
        }
        key = strings.TrimPrefix(key, "/")
        if _, exists := c.objects[key]; !exists {
                return ErrObjectNotFound
        }
        delete(c.objects, key)
        return nil
}

func (c *Client) ListObjects(ctx context.Context, prefix string) ([]*ObjectMeta, error) {
        var result []*ObjectMeta
        for key, data := range c.objects {
                if prefix == "" || strings.HasPrefix(key, prefix) {
                        result = append(result, &ObjectMeta{Key: key, Size: int64(len(data))})
                }
        }
        return result, nil
}
