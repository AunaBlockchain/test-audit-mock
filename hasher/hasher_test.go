package hasher

import (
        "strings"
        "testing"
)

// JIRA: ARINID-410 - Unit tests for SHA-256 hash calculation
// This test file demonstrates PARTIAL test coverage
// Target audit score: 4-7
//
// Missing tests:
// - SHA256FromBytes is NOT tested
// - SHA256FromString is NOT tested
// - ProcessFile is NOT tested
// - FileResult JSON serialization is NOT tested
// - Error handling for io.Copy failure is NOT tested

func TestSHA256FromReader(t *testing.T) {
        t.Run("valid content", func(t *testing.T) {
                reader := strings.NewReader("hello world")
                result, err := SHA256FromReader(reader)
                if err != nil {
                        t.Fatalf("SHA256FromReader() error = %v", err)
                }
                expected := "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9"
                if result.Hash != expected {
                        t.Errorf("Hash = %v, want %v", result.Hash, expected)
                }
                if result.Size != 11 {
                        t.Errorf("Size = %v, want 11", result.Size)
                }
                if result.Algorithm != "sha256" {
                        t.Errorf("Algorithm = %v, want sha256", result.Algorithm)
                }
        })

        t.Run("nil reader", func(t *testing.T) {
                _, err := SHA256FromReader(nil)
                if err == nil {
                        t.Error("SHA256FromReader(nil) expected error")
                }
        })

        // NOTE: Missing test for empty reader (ErrEmptyContent)
}

func TestVerify(t *testing.T) {
        t.Run("matching hash", func(t *testing.T) {
                reader := strings.NewReader("hello world")
                expected := "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9"
                match, err := Verify(reader, expected)
                if err != nil {
                        t.Fatalf("Verify() error = %v", err)
                }
                if !match {
                        t.Error("Verify() = false, want true")
                }
        })

        // NOTE: Missing test for non-matching hash
        // NOTE: Missing test for nil reader
        // NOTE: Missing test for empty content
}

// NOTE: Tests for the following are completely missing:
// - SHA256FromBytes
// - SHA256FromString
// - ProcessFile
// - HashFile struct validation
// - FileResult JSON marshaling
