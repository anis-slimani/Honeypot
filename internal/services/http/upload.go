package http

import (
	"crypto/md5"
	"crypto/sha256"
	"database/sql"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"honey/internal/config"
	"honey/internal/database"
	"honey/internal/logger"
	"honey/internal/models"
)

// UploadApp handles file upload honeypot
type UploadApp struct {
	config   *config.HTTPConfig
	logger   logger.Logger
	db       *sql.DB
	detector *VulnerabilityDetector
}

// NewUploadApp creates a new upload honeypot
func NewUploadApp(cfg *config.HTTPConfig, log logger.Logger, db *sql.DB, detector *VulnerabilityDetector) *UploadApp {
	return &UploadApp{
		config:   cfg,
		logger:   log,
		db:       db,
		detector: detector,
	}
}

// Paths returns handled paths
func (u *UploadApp) Paths() []string {
	paths := []string{}
	for _, p := range u.config.Applications.Upload.Paths {
		paths = append(paths, p)
	}
	return paths
}

// HandleRequest handles upload requests
func (u *UploadApp) HandleRequest(r *http.Request, body string, requestID int) *Response {
	// Run vulnerability detection
	u.detector.Analyze(r, r.URL.RawQuery+" "+body, requestID)

	if r.Method == "POST" {
		return u.handleUpload(r, requestID)
	}

	return &Response{
		StatusCode: 200,
		Headers:    map[string][]string{"Content-Type": {"text/html"}},
		Body:       u.generateUploadPage(),
	}
}

// handleUpload handles file uploads
func (u *UploadApp) handleUpload(r *http.Request, requestID int) *Response {
	// Parse multipart form
	err := r.ParseMultipartForm(u.config.Applications.Upload.MaxFileSize)
	if err != nil {
		u.logger.Errorf("[Upload] Failed to parse multipart form: %v", err)
		return &Response{
			StatusCode: 400,
			Headers:    map[string][]string{"Content-Type": {"application/json"}},
			Body:       `{"error": "Failed to parse upload"}`,
		}
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		u.logger.Errorf("[Upload] Failed to get file: %v", err)
		return &Response{
			StatusCode: 400,
			Headers:    map[string][]string{"Content-Type": {"application/json"}},
			Body:       `{"error": "No file provided"}`,
		}
	}
	defer file.Close()

	return u.saveUploadedFile(file, header, requestID)
}

// saveUploadedFile saves an uploaded file
func (u *UploadApp) saveUploadedFile(file multipart.File, header *multipart.FileHeader, requestID int) *Response {
	// Read file content
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		u.logger.Errorf("[Upload] Failed to read file: %v", err)
		return &Response{
			StatusCode: 500,
			Headers:    map[string][]string{"Content-Type": {"application/json"}},
			Body:       `{"error": "Failed to read file"}`,
		}
	}

	// Generate hashes
	md5Hash := fmt.Sprintf("%x", md5.Sum(fileBytes))
	sha256Hash := fmt.Sprintf("%x", sha256.Sum256(fileBytes))

	// Create quarantine directory if it doesn't exist
	quarantineDir := u.config.Applications.Upload.QuarantineDir
	os.MkdirAll(quarantineDir, 0755)

	// Generate safe filename
	safeFilename := fmt.Sprintf("%d_%s", time.Now().Unix(), filepath.Base(header.Filename))
	filePath := filepath.Join(quarantineDir, safeFilename)

	// Save file
	err = os.WriteFile(filePath, fileBytes, 0644)
	if err != nil {
		u.logger.Errorf("[Upload] Failed to save file: %v", err)
	}

	// Detect malicious file
	isMalicious := u.detectMaliciousFile(header.Filename, fileBytes)

	u.logger.Warnf("[Upload] File uploaded: %s (MD5: %s, Malicious: %v)", header.Filename, md5Hash, isMalicious)

	// Save to database
	uploadedFile := &models.UploadedFile{
		RequestID:        requestID,
		Filename:         safeFilename,
		OriginalFilename: header.Filename,
		Size:             int64(len(fileBytes)),
		ContentType:      header.Header.Get("Content-Type"),
		MD5Hash:          md5Hash,
		SHA256Hash:       sha256Hash,
		FilePath:         filePath,
		IsMalicious:      isMalicious,
		Timestamp:        time.Now(),
	}

	database.SaveUploadedFile(uploadedFile)

	// Return fake success response
	return &Response{
		StatusCode: 200,
		Headers:    map[string][]string{"Content-Type": {"application/json"}},
		Body: fmt.Sprintf(`{
  "success": true,
  "filename": "%s",
  "path": "/var/www/uploads/%s",
  "url": "http://localhost/uploads/%s",
  "size": %d
}`, header.Filename, header.Filename, header.Filename, len(fileBytes)),
	}
}

// detectMaliciousFile detects potentially malicious files
func (u *UploadApp) detectMaliciousFile(filename string, content []byte) bool {
	// Check file extension
	maliciousExtensions := []string{
		".php", ".phtml", ".php3", ".php4", ".php5", ".php7",
		".asp", ".aspx", ".jsp", ".jspx",
		".sh", ".bash", ".zsh",
		".exe", ".bat", ".cmd", ".com",
		".pl", ".py", ".rb",
	}

	lowerFilename := strings.ToLower(filename)
	for _, ext := range maliciousExtensions {
		if strings.HasSuffix(lowerFilename, ext) {
			return true
		}
	}

	// Check for double extensions
	if strings.Count(filename, ".") > 1 {
		parts := strings.Split(filename, ".")
		if len(parts) >= 2 {
			for _, ext := range maliciousExtensions {
				if "."+parts[len(parts)-2] == ext {
					return true
				}
			}
		}
	}

	// Check content for common web shell signatures
	contentStr := string(content)
	webshellSignatures := []string{
		"<?php",
		"eval(",
		"base64_decode",
		"system(",
		"exec(",
		"shell_exec(",
		"passthru(",
		"popen(",
		"proc_open(",
		"<?=",
	}

	for _, sig := range webshellSignatures {
		if strings.Contains(contentStr, sig) {
			return true
		}
	}

	return false
}

// generateUploadPage generates an upload form
func (u *UploadApp) generateUploadPage() string {
	return `<!DOCTYPE html>
<html>
<head>
    <title>File Upload</title>
    <style>
        body { font-family: Arial; padding: 50px; background: #f0f0f0; }
        .upload-box { background: white; padding: 30px; border-radius: 5px; max-width: 500px; margin: 0 auto; }
        input[type=file] { margin: 20px 0; }
        button { padding: 10px 20px; background: #4CAF50; color: white; border: none; border-radius: 3px; cursor: pointer; }
    </style>
</head>
<body>
    <div class="upload-box">
        <h2>File Upload</h2>
        <form method="post" enctype="multipart/form-data">
            <input type="file" name="file" required>
            <br>
            <button type="submit">Upload</button>
        </form>
    </div>
</body>
</html>`
}
