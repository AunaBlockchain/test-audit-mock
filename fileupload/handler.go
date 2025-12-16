// Package fileupload provides HTTP handlers for file upload operations.
// JIRA: ARINID-409 - Implement POST /files endpoint for PDF upload to Garage
//
// Acceptance Criteria:
// - Endpoint POST /files
// - PDF MIME validation
// - Send file to S3 bucket
// - Error handling and timeout
// - Return JSON with { uri, sha256 }
package fileupload

import (
        "encoding/json"
        "errors"
        "fmt"
        "io"
        "mime/multipart"
        "net/http"
        "path/filepath"
        "strings"
        "time"
)

var (
        ErrInvalidMIME = errors.New("invalid MIME type: only PDF files are allowed")
        ErrFileTooLarge = errors.New("file too large")
        ErrMissingFile = errors.New("no file provided")
)

const MaxFileSize = 10 * 1024 * 1024

var AllowedMIMETypes = []string{"application/pdf"}

type UploadResponse struct {
        URI        string    `json:"uri"`
        SHA256     string    `json:"sha256"`
        FileName   string    `json:"fileName"`
        Size       int64     `json:"size"`
        UploadedAt time.Time `json:"uploadedAt"`
}

type ErrorResponse struct {
        Error   string `json:"error"`
        Code    string `json:"code"`
        Message string `json:"message"`
}

type Handler struct {
        maxSize     int64
        allowedMIME []string
        uploader    Uploader
}

type Uploader interface {
        Upload(filename string, data io.Reader, contentType string) (uri string, sha256 string, err error)
}

func NewHandler(uploader Uploader) *Handler {
        return &Handler{maxSize: MaxFileSize, allowedMIME: AllowedMIMETypes, uploader: uploader}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodPost {
                h.writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST method is allowed")
                return
        }
        if err := r.ParseMultipartForm(h.maxSize); err != nil {
                h.writeError(w, http.StatusBadRequest, "PARSE_ERROR", "Failed to parse form")
                return
        }
        file, header, err := r.FormFile("file")
        if err != nil {
                h.writeError(w, http.StatusBadRequest, "MISSING_FILE", "No file provided")
                return
        }
        defer file.Close()
        if header.Size > h.maxSize {
                h.writeError(w, http.StatusRequestEntityTooLarge, "FILE_TOO_LARGE", "File exceeds maximum size")
                return
        }
        if !h.isValidMIME(header) {
                h.writeError(w, http.StatusUnsupportedMediaType, "INVALID_MIME", "Only PDF files are allowed")
                return
        }
        uri, sha256, err := h.uploader.Upload(header.Filename, file, header.Header.Get("Content-Type"))
        if err != nil {
                h.writeError(w, http.StatusInternalServerError, "UPLOAD_ERROR", "Failed to upload file")
                return
        }
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusCreated)
        json.NewEncoder(w).Encode(UploadResponse{URI: uri, SHA256: sha256, FileName: header.Filename, Size: header.Size, UploadedAt: time.Now()})
}

func (h *Handler) isValidMIME(header *multipart.FileHeader) bool {
        contentType := header.Header.Get("Content-Type")
        for _, allowed := range h.allowedMIME {
                if strings.HasPrefix(contentType, allowed) {
                        return true
                }
        }
        return strings.ToLower(filepath.Ext(header.Filename)) == ".pdf"
}

func (h *Handler) writeError(w http.ResponseWriter, status int, code, message string) {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(status)
        json.NewEncoder(w).Encode(ErrorResponse{Error: http.StatusText(status), Code: code, Message: message})
}

func ValidatePDF(file multipart.File) error {
        header := make([]byte, 5)
        if _, err := file.Read(header); err != nil {
                return fmt.Errorf("failed to read file header: %w", err)
        }
        if _, err := file.Seek(0, 0); err != nil {
                return fmt.Errorf("failed to seek file: %w", err)
        }
        if string(header) != "%PDF-" {
                return ErrInvalidMIME
        }
        return nil
}

func GenerateURI(bucket, key string) string {
        return fmt.Sprintf("s3://garage/%s/%s", bucket, key)
}

func ParseURI(uri string) (bucket, key string, err error) {
        if !strings.HasPrefix(uri, "s3://garage/") {
                return "", "", errors.New("invalid URI format")
        }
        path := strings.TrimPrefix(uri, "s3://garage/")
        parts := strings.SplitN(path, "/", 2)
        if len(parts) != 2 {
                return "", "", errors.New("invalid URI format: missing key")
        }
        return parts[0], parts[1], nil
}
