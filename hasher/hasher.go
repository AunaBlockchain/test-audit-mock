// Package hasher provides cryptographic hashing utilities.
// JIRA: ARINID-410 - Implement SHA-256 calculation for PDF files
//
// Acceptance Criteria:
// - Function to calculate hash from io.Reader
// - Integration with upload
// - Return hash in hex string format
package hasher

import (
        "crypto/sha256"
        "encoding/hex"
        "errors"
        "fmt"
        "io"
)

var (
        ErrNilReader    = errors.New("reader cannot be nil")
        ErrEmptyContent = errors.New("content cannot be empty")
)

type Result struct {
        Hash      string
        Size      int64
        Algorithm string
}

func SHA256FromReader(r io.Reader) (*Result, error) {
        if r == nil {
                return nil, ErrNilReader
        }
        hash := sha256.New()
        n, err := io.Copy(hash, r)
        if err != nil {
                return nil, fmt.Errorf("failed to read content: %w", err)
        }
        if n == 0 {
                return nil, ErrEmptyContent
        }
        return &Result{Hash: hex.EncodeToString(hash.Sum(nil)), Size: n, Algorithm: "sha256"}, nil
}

func SHA256FromBytes(data []byte) (*Result, error) {
        if len(data) == 0 {
                return nil, ErrEmptyContent
        }
        hash := sha256.Sum256(data)
        return &Result{Hash: hex.EncodeToString(hash[:]), Size: int64(len(data)), Algorithm: "sha256"}, nil
}

func SHA256FromString(s string) (*Result, error) {
        return SHA256FromBytes([]byte(s))
}

func Verify(reader io.Reader, expectedHash string) (bool, error) {
        result, err := SHA256FromReader(reader)
        if err != nil {
                return false, err
        }
        return result.Hash == expectedHash, nil
}

type HashFile struct {
        Name     string
        MimeType string
        Data     io.Reader
}

type FileResult struct {
        Name     string `json:"name"`
        MimeType string `json:"mimeType"`
        Hash     string `json:"hash"`
        Size     int64  `json:"size"`
}

func ProcessFile(file *HashFile) (*FileResult, error) {
        if file == nil {
                return nil, errors.New("file cannot be nil")
        }
        if file.Data == nil {
                return nil, ErrNilReader
        }
        result, err := SHA256FromReader(file.Data)
        if err != nil {
                return nil, err
        }
        return &FileResult{Name: file.Name, MimeType: file.MimeType, Hash: result.Hash, Size: result.Size}, nil
}
