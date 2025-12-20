package models

import (
	"time"
)

// HTTPRequest represents an HTTP request captured by the honeypot
type HTTPRequest struct {
	ID           int       `json:"id"`
	RemoteAddr   string    `json:"remote_addr"`
	Method       string    `json:"method"`
	Path         string    `json:"path"`
	QueryString  string    `json:"query_string,omitempty"`
	UserAgent    string    `json:"user_agent,omitempty"`
	Referer      string    `json:"referer,omitempty"`
	Headers      string    `json:"headers,omitempty"` // JSON encoded
	Body         string    `json:"body,omitempty"`
	Timestamp    time.Time `json:"timestamp"`
	Country      string    `json:"country,omitempty"`
	City         string    `json:"city,omitempty"`
	ResponseCode int       `json:"response_code"`
}

// HTTPAttack represents a detected HTTP attack
type HTTPAttack struct {
	ID               int       `json:"id"`
	RequestID        int       `json:"request_id"`
	AttackType       string    `json:"attack_type"` // sqli, xss, rfi, lfi, cmd_injection, etc.
	Severity         string    `json:"severity"`    // low, medium, high, critical
	Payload          string    `json:"payload"`
	MatchedPattern   string    `json:"matched_pattern,omitempty"`
	ResponseStrategy string    `json:"response_strategy,omitempty"`
	Timestamp        time.Time `json:"timestamp"`
}

// UploadedFile represents a file uploaded to the honeypot
type UploadedFile struct {
	ID               int        `json:"id"`
	RequestID        int        `json:"request_id"`
	Filename         string     `json:"filename"`
	OriginalFilename string     `json:"original_filename"`
	Size             int64      `json:"size"`
	ContentType      string     `json:"content_type"`
	MD5Hash          string     `json:"md5_hash"`
	SHA256Hash       string     `json:"sha256_hash"`
	FilePath         string     `json:"file_path"`
	IsMalicious      bool       `json:"is_malicious"`
	VirusTotalResult string     `json:"virustotal_result,omitempty"` // JSON encoded
	Timestamp        time.Time  `json:"timestamp"`
}

// ScannerDetection represents a detected automated scanner
type ScannerDetection struct {
	ID                int       `json:"id"`
	RemoteAddr        string    `json:"remote_addr"`
	ScannerType       string    `json:"scanner_type"` // sqlmap, nikto, nmap, etc.
	Version           string    `json:"version,omitempty"`
	RequestCount      int       `json:"request_count"`
	RequestsPerSecond float64   `json:"requests_per_second"`
	DetectedAt        time.Time `json:"detected_at"`
	LastSeen          time.Time `json:"last_seen"`
}

// HTTPCredential represents login credentials captured
type HTTPCredential struct {
	ID          int       `json:"id"`
	RequestID   int       `json:"request_id"`
	Application string    `json:"application"` // wordpress, phpmyadmin, admin, etc.
	Username    string    `json:"username"`
	Password    string    `json:"password"`
	Success     bool      `json:"success"`
	Timestamp   time.Time `json:"timestamp"`
}

// HTTPStatistics represents HTTP-specific statistics
type HTTPStatistics struct {
	TotalRequests       int                    `json:"total_requests"`
	TotalAttacks        int                    `json:"total_attacks"`
	TotalUploads        int                    `json:"total_uploads"`
	MaliciousUploads    int                    `json:"malicious_uploads"`
	TopAttackTypes      []AttackTypeStats      `json:"top_attack_types"`
	TopPaths            []PathStats            `json:"top_paths"`
	TopUserAgents       []UserAgentStats       `json:"top_user_agents"`
	DetectedScanners    []ScannerDetection     `json:"detected_scanners"`
	CredentialAttempts  []CredentialStats      `json:"credential_attempts"`
}

// AttackTypeStats statistics by attack type
type AttackTypeStats struct {
	AttackType string `json:"attack_type"`
	Count      int    `json:"count"`
}

// PathStats statistics by request path
type PathStats struct {
	Path  string `json:"path"`
	Count int    `json:"count"`
}

// UserAgentStats statistics by user agent
type UserAgentStats struct {
	UserAgent string `json:"user_agent"`
	Count     int    `json:"count"`
}

// CredentialStats statistics for credential attempts
type CredentialStats struct {
	Application string `json:"application"`
	Username    string `json:"username"`
	Count       int    `json:"count"`
}
