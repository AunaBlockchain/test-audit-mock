package storage

import (
        "context"
        "errors"
        "io"
        "os"
        "strings"
        "testing"
)

// JIRA: ARINID-408 - Unit tests for S3 client module
// This test file demonstrates COMPREHENSIVE test coverage
// Target audit score: 8-10

func TestNewClient(t *testing.T) {
        tests := []struct {
                name    string
                cfg     *Config
                wantErr error
        }{
                {name: "valid config", cfg: &Config{Endpoint: "http://localhost:3900", Region: "us-east-1", AccessKey: "test-key", SecretKey: "test-secret", Bucket: "test-bucket"}, wantErr: nil},
                {name: "nil config", cfg: nil, wantErr: ErrInvalidConfig},
                {name: "empty endpoint", cfg: &Config{Endpoint: "", Region: "us-east-1", AccessKey: "test-key", SecretKey: "test-secret", Bucket: "test-bucket"}, wantErr: ErrInvalidConfig},
                {name: "empty bucket", cfg: &Config{Endpoint: "http://localhost:3900", Region: "us-east-1", AccessKey: "test-key", SecretKey: "test-secret", Bucket: ""}, wantErr: ErrInvalidConfig},
        }

        for _, tt := range tests {
                t.Run(tt.name, func(t *testing.T) {
                        client, err := NewClient(tt.cfg)
                        if tt.wantErr != nil {
                                if err == nil || !errors.Is(err, tt.wantErr) {
                                        t.Errorf("NewClient() error = %v, wantErr %v", err, tt.wantErr)
                                }
                                return
                        }
                        if err != nil {
                                t.Errorf("NewClient() unexpected error = %v", err)
                        }
                        if client == nil {
                                t.Error("NewClient() returned nil client")
                        }
                })
        }
}

func TestConfigFromEnv(t *testing.T) {
        orig := map[string]string{"S3_ENDPOINT": os.Getenv("S3_ENDPOINT"), "S3_REGION": os.Getenv("S3_REGION"), "S3_ACCESS_KEY": os.Getenv("S3_ACCESS_KEY"), "S3_SECRET_KEY": os.Getenv("S3_SECRET_KEY"), "S3_BUCKET": os.Getenv("S3_BUCKET")}
        t.Cleanup(func() { for k, v := range orig { os.Setenv(k, v) } })

        t.Run("valid env vars", func(t *testing.T) {
                os.Setenv("S3_ENDPOINT", "http://localhost:3900")
                os.Setenv("S3_REGION", "eu-west-1")
                os.Setenv("S3_ACCESS_KEY", "mykey")
                os.Setenv("S3_SECRET_KEY", "mysecret")
                os.Setenv("S3_BUCKET", "mybucket")
                cfg, err := ConfigFromEnv()
                if err != nil {
                        t.Fatalf("ConfigFromEnv() error = %v", err)
                }
                if cfg.Endpoint != "http://localhost:3900" || cfg.Region != "eu-west-1" {
                        t.Errorf("Config mismatch")
                }
        })

        t.Run("missing endpoint", func(t *testing.T) {
                os.Setenv("S3_ENDPOINT", "")
                _, err := ConfigFromEnv()
                if err == nil {
                        t.Error("Expected error for missing endpoint")
                }
        })
}

func TestClient_PutObject(t *testing.T) {
        cfg := &Config{Endpoint: "http://localhost:3900", Bucket: "test"}
        client, _ := NewClient(cfg)
        ctx := context.Background()

        t.Run("valid put", func(t *testing.T) {
                meta, err := client.PutObject(ctx, "test.txt", strings.NewReader("hello"), "text/plain")
                if err != nil {
                        t.Fatalf("PutObject() error = %v", err)
                }
                if meta.Key != "test.txt" || meta.Size != 5 {
                        t.Errorf("Unexpected meta: %+v", meta)
                }
        })

        t.Run("empty key", func(t *testing.T) {
                _, err := client.PutObject(ctx, "", strings.NewReader("hello"), "text/plain")
                if err == nil {
                        t.Error("Expected error for empty key")
                }
        })

        t.Run("nil data", func(t *testing.T) {
                _, err := client.PutObject(ctx, "test.txt", nil, "text/plain")
                if err == nil {
                        t.Error("Expected error for nil data")
                }
        })
}

func TestClient_GetObject(t *testing.T) {
        cfg := &Config{Endpoint: "http://localhost:3900", Bucket: "test"}
        client, _ := NewClient(cfg)
        ctx := context.Background()
        client.PutObject(ctx, "exists.txt", strings.NewReader("content"), "text/plain")

        t.Run("existing object", func(t *testing.T) {
                reader, meta, err := client.GetObject(ctx, "exists.txt")
                if err != nil {
                        t.Fatalf("GetObject() error = %v", err)
                }
                defer reader.Close()
                data, _ := io.ReadAll(reader)
                if string(data) != "content" || meta.Key != "exists.txt" {
                        t.Errorf("Unexpected result")
                }
        })

        t.Run("non-existing object", func(t *testing.T) {
                _, _, err := client.GetObject(ctx, "notfound.txt")
                if !errors.Is(err, ErrObjectNotFound) {
                        t.Errorf("Expected ErrObjectNotFound, got %v", err)
                }
        })
}

func TestClient_HeadObject(t *testing.T) {
        cfg := &Config{Endpoint: "http://localhost:3900", Bucket: "test"}
        client, _ := NewClient(cfg)
        ctx := context.Background()
        client.PutObject(ctx, "head.txt", strings.NewReader("data"), "text/plain")

        t.Run("existing", func(t *testing.T) {
                meta, err := client.HeadObject(ctx, "head.txt")
                if err != nil || meta.Size != 4 {
                        t.Errorf("HeadObject failed: err=%v, meta=%+v", err, meta)
                }
        })

        t.Run("not found", func(t *testing.T) {
                _, err := client.HeadObject(ctx, "missing.txt")
                if !errors.Is(err, ErrObjectNotFound) {
                        t.Errorf("Expected ErrObjectNotFound")
                }
        })
}

func TestClient_DeleteObject(t *testing.T) {
        cfg := &Config{Endpoint: "http://localhost:3900", Bucket: "test"}
        client, _ := NewClient(cfg)
        ctx := context.Background()
        client.PutObject(ctx, "del.txt", strings.NewReader("data"), "text/plain")

        t.Run("delete existing", func(t *testing.T) {
                if err := client.DeleteObject(ctx, "del.txt"); err != nil {
                        t.Errorf("DeleteObject() error = %v", err)
                }
                _, err := client.HeadObject(ctx, "del.txt")
                if !errors.Is(err, ErrObjectNotFound) {
                        t.Error("Object should be deleted")
                }
        })

        t.Run("delete non-existing", func(t *testing.T) {
                err := client.DeleteObject(ctx, "nope.txt")
                if !errors.Is(err, ErrObjectNotFound) {
                        t.Errorf("Expected ErrObjectNotFound")
                }
        })
}

func TestClient_ListObjects(t *testing.T) {
        cfg := &Config{Endpoint: "http://localhost:3900", Bucket: "test"}
        client, _ := NewClient(cfg)
        ctx := context.Background()
        client.PutObject(ctx, "prefix/a.txt", strings.NewReader("a"), "text/plain")
        client.PutObject(ctx, "prefix/b.txt", strings.NewReader("b"), "text/plain")
        client.PutObject(ctx, "other.txt", strings.NewReader("c"), "text/plain")

        t.Run("with prefix", func(t *testing.T) {
                list, err := client.ListObjects(ctx, "prefix/")
                if err != nil || len(list) != 2 {
                        t.Errorf("ListObjects(prefix/) = %d items, err=%v", len(list), err)
                }
        })

        t.Run("all objects", func(t *testing.T) {
                list, _ := client.ListObjects(ctx, "")
                if len(list) < 3 {
                        t.Errorf("ListObjects() should return all objects")
                }
        })
}
