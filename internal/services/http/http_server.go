package http

import (
	"context"
	"crypto/md5"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"honey/internal/config"
	"honey/internal/database"
	"honey/internal/logger"
	"honey/internal/models"
)

// HTTPHoneypot represents the HTTP honeypot server
type HTTPHoneypot struct {
	config        *config.HTTPConfig
	logger        logger.Logger
	db            *sql.DB
	server        *http.Server
	router        *Router
	mu            sync.RWMutex
	requestCounts map[string]*RequestCount
}

// RequestCount tracks requests per IP for scanner detection
type RequestCount struct {
	Count      int
	FirstSeen  time.Time
	LastSeen   time.Time
	UserAgent  string
}

// New creates a new HTTP honeypot instance
func New(cfg *config.HTTPConfig, log logger.Logger, db *sql.DB) *HTTPHoneypot {
	hp := &HTTPHoneypot{
		config:        cfg,
		logger:        log,
		db:            db,
		router:        NewRouter(cfg, log, db),
		requestCounts: make(map[string]*RequestCount),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", hp.handleRequest)

	hp.server = &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return hp
}

// Start starts the HTTP honeypot server
func (h *HTTPHoneypot) Start(ctx context.Context) error {
	h.logger.Infof("Starting HTTP honeypot on %s", h.server.Addr)

	errChan := make(chan error, 1)

	go func() {
		if h.config.TLS.Enabled {
			h.logger.Info("Starting HTTPS server...")
			if err := h.server.ListenAndServeTLS(h.config.TLS.CertFile, h.config.TLS.KeyFile); err != nil && err != http.ErrServerClosed {
				errChan <- fmt.Errorf("HTTPS server error: %w", err)
			}
		} else {
			if err := h.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				errChan <- fmt.Errorf("HTTP server error: %w", err)
			}
		}
	}()

	// Wait for context cancellation or error
	select {
	case <-ctx.Done():
		h.logger.Info("Shutting down HTTP honeypot...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return h.server.Shutdown(shutdownCtx)
	case err := <-errChan:
		return err
	}
}

// Stop stops the HTTP honeypot server
func (h *HTTPHoneypot) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return h.server.Shutdown(ctx)
}

// handleRequest is the main request handler
func (h *HTTPHoneypot) handleRequest(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	// Get client IP
	clientIP := h.getClientIP(r)

	h.logger.Infof("[HTTP] %s %s %s from %s", r.Method, r.URL.Path, r.Proto, clientIP)

	// Track request for scanner detection
	h.trackRequest(clientIP, r.UserAgent())

	// Read request body (except for multipart uploads)
	var body string
	contentType := r.Header.Get("Content-Type")
	if !strings.Contains(contentType, "multipart/form-data") {
		bodyBytes, _ := io.ReadAll(r.Body)
		r.Body.Close()
		body = string(bodyBytes)
	}

	// Capture headers
	headersJSON, _ := json.Marshal(r.Header)

	// Create HTTP request record
	httpReq := &models.HTTPRequest{
		RemoteAddr:  clientIP,
		Method:      r.Method,
		Path:        r.URL.Path,
		QueryString: r.URL.RawQuery,
		UserAgent:   r.UserAgent(),
		Referer:     r.Referer(),
		Headers:     string(headersJSON),
		Body:        body,
		Timestamp:   time.Now(),
	}

	// Save request to database
	requestID, err := database.SaveHTTPRequest(httpReq)
	if err != nil {
		h.logger.Errorf("Failed to save HTTP request: %v", err)
	}

	// Detect scanner
	h.detectScanner(clientIP, r.UserAgent(), requestID)

	// Route the request through the virtual host router
	response := h.router.Route(r, body, requestID)

	// Update response code in database
	httpReq.ResponseCode = response.StatusCode
	database.UpdateHTTPRequestResponse(requestID, response.StatusCode)

	// Write response
	for key, values := range response.Headers {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}
	w.WriteHeader(response.StatusCode)
	w.Write([]byte(response.Body))

	// Log response time
	duration := time.Since(startTime)
	h.logger.Debugf("[HTTP] Request handled in %v", duration)
}

// getClientIP extracts the client IP from the request
func (h *HTTPHoneypot) getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[0])
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	ip := r.RemoteAddr
	if colon := strings.LastIndex(ip, ":"); colon != -1 {
		ip = ip[:colon]
	}
	return ip
}

// trackRequest tracks requests per IP for scanner detection
func (h *HTTPHoneypot) trackRequest(ip, userAgent string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if count, exists := h.requestCounts[ip]; exists {
		count.Count++
		count.LastSeen = time.Now()
		if count.UserAgent == "" {
			count.UserAgent = userAgent
		}
	} else {
		h.requestCounts[ip] = &RequestCount{
			Count:     1,
			FirstSeen: time.Now(),
			LastSeen:  time.Now(),
			UserAgent: userAgent,
		}
	}
}

// detectScanner detects automated scanners based on user agent and request patterns
func (h *HTTPHoneypot) detectScanner(ip, userAgent string, requestID int) {
	h.mu.RLock()
	count := h.requestCounts[ip]
	h.mu.RUnlock()

	if count == nil {
		return
	}

	// Detect scanner by User-Agent
	scannerType := ""
	scannerSignatures := map[string]string{
		"sqlmap":     "sqlmap",
		"nikto":      "Nikto",
		"nmap":       "Nmap",
		"masscan":    "masscan",
		"zap":        "OWASP ZAP",
		"burp":       "Burp",
		"acunetix":   "Acunetix",
		"nessus":     "Nessus",
		"metasploit": "Metasploit",
		"nuclei":     "Nuclei",
		"wpscan":     "WPScan",
		"gobuster":   "gobuster",
		"dirbuster":  "DirBuster",
	}

	lowerUA := strings.ToLower(userAgent)
	for key, name := range scannerSignatures {
		if strings.Contains(lowerUA, key) {
			scannerType = name
			break
		}
	}

	// Detect scanner by request speed
	if count.Count > 10 {
		duration := count.LastSeen.Sub(count.FirstSeen).Seconds()
		if duration > 0 {
			rps := float64(count.Count) / duration

			// If more than 2 requests per second, likely automated
			if rps > 2.0 {
				if scannerType == "" {
					scannerType = "Unknown Scanner"
				}

				// Save scanner detection
				scanner := &models.ScannerDetection{
					RemoteAddr:        ip,
					ScannerType:       scannerType,
					RequestCount:      count.Count,
					RequestsPerSecond: rps,
					DetectedAt:        count.FirstSeen,
					LastSeen:          count.LastSeen,
				}

				if err := database.SaveScannerDetection(scanner); err != nil {
					h.logger.Errorf("Failed to save scanner detection: %v", err)
				} else {
					h.logger.Warnf("[SCANNER DETECTED] %s from %s (%.2f req/s)", scannerType, ip, rps)
				}
			}
		}
	}
}

// hashFile generates MD5 and SHA256 hashes for a file
func hashFile(data []byte) (string, string) {
	md5Hash := md5.Sum(data)
	sha256Hash := sha256.Sum256(data)
	return fmt.Sprintf("%x", md5Hash), fmt.Sprintf("%x", sha256Hash)
}
