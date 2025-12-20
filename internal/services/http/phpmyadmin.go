package http

import (
	"database/sql"
	"fmt"
	"net/http"
	"strings"
	"time"

	"honey/internal/config"
	"honey/internal/database"
	"honey/internal/logger"
	"honey/internal/models"
)

// PHPMyAdminApp emulates phpMyAdmin
type PHPMyAdminApp struct {
	config   *config.HTTPConfig
	logger   logger.Logger
	db       *sql.DB
	detector *VulnerabilityDetector
}

// NewPHPMyAdminApp creates a new phpMyAdmin honeypot
func NewPHPMyAdminApp(cfg *config.HTTPConfig, log logger.Logger, db *sql.DB, detector *VulnerabilityDetector) *PHPMyAdminApp {
	return &PHPMyAdminApp{
		config:   cfg,
		logger:   log,
		db:       db,
		detector: detector,
	}
}

// Paths returns handled paths
func (p *PHPMyAdminApp) Paths() []string {
	base := p.config.Applications.PHPMyAdmin.Path
	return []string{
		base,
		base + "/",
		base + "/index.php",
		base + "/setup.php",
		base + "/sql.php",
		base + "/import.php",
		base + "/export.php",
	}
}

// HandleRequest handles phpMyAdmin requests
func (p *PHPMyAdminApp) HandleRequest(r *http.Request, body string, requestID int) *Response {
	path := r.URL.Path

	// Run vulnerability detection
	p.detector.Analyze(r, r.URL.RawQuery+" "+body, requestID)

	if strings.HasSuffix(path, "/setup.php") {
		p.logger.Warnf("[phpMyAdmin] Setup.php access attempt (CVE-2018-12613)")
		return p.handleSetup(r, body, requestID)
	}

	if r.Method == "POST" {
		return p.handleLogin(r, body, requestID)
	}

	return &Response{
		StatusCode: 200,
		Headers:    map[string][]string{"Content-Type": {"text/html"}},
		Body:       p.generateLoginPage(false),
	}
}

// handleLogin handles login attempts
func (p *PHPMyAdminApp) handleLogin(r *http.Request, body string, requestID int) *Response {
	username := ""
	password := ""

	parts := strings.Split(body, "&")
	for _, part := range parts {
		kv := strings.SplitN(part, "=", 2)
		if len(kv) == 2 {
			if kv[0] == "pma_username" {
				username = kv[1]
			}
			if kv[0] == "pma_password" {
				password = kv[1]
			}
		}
	}

	p.logger.Infof("[phpMyAdmin] Login attempt - Username: %s, Password: %s", username, password)

	cred := &models.HTTPCredential{
		RequestID:   requestID,
		Application: "phpmyadmin",
		Username:    username,
		Password:    password,
		Success:     false,
		Timestamp:   time.Now(),
	}

	// Check fake credentials
	for _, user := range p.config.Applications.PHPMyAdmin.FakeCredentials {
		if username == user.Username && password == user.Password {
			cred.Success = true
			break
		}
	}

	database.SaveHTTPCredential(cred)

	if cred.Success {
		return &Response{
			StatusCode: 200,
			Headers:    map[string][]string{"Content-Type": {"text/html"}},
			Body:       p.generateDashboard(),
		}
	}

	return &Response{
		StatusCode: 200,
		Headers:    map[string][]string{"Content-Type": {"text/html"}},
		Body:       p.generateLoginPage(true),
	}
}

// handleSetup handles setup.php (vulnerable endpoint)
func (p *PHPMyAdminApp) handleSetup(r *http.Request, body string, requestID int) *Response {
	return &Response{
		StatusCode: 200,
		Headers:    map[string][]string{"Content-Type": {"text/html"}},
		Body: `<!DOCTYPE html>
<html>
<head><title>phpMyAdmin setup</title></head>
<body>
<h1>phpMyAdmin ` + p.config.Applications.PHPMyAdmin.Version + ` - Setup</h1>
<p>Configuration file now exists, ignoring setup parameters.</p>
<!-- CVE-2018-12613: Remote Code Execution vulnerability -->
</body>
</html>`,
	}
}

// generateLoginPage generates phpMyAdmin login page
func (p *PHPMyAdminApp) generateLoginPage(failed bool) string {
	errorMsg := ""
	if failed {
		errorMsg = `<div class="error">Access denied!</div>`
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <title>phpMyAdmin</title>
    <style>
        body { font-family: sans-serif; background: #f5f5f5; }
        .login { width: 400px; margin: 100px auto; background: white; padding: 30px; box-shadow: 0 0 10px rgba(0,0,0,0.1); }
        input { width: 100%%; padding: 10px; margin: 5px 0; box-sizing: border-box; }
        input[type=submit] { background: #007bff; color: white; border: none; cursor: pointer; }
        .error { color: red; padding: 10px; background: #ffebee; margin-bottom: 10px; }
    </style>
</head>
<body>
    <div class="login">
        <h2>phpMyAdmin %s</h2>
        %s
        <form method="post">
            <label>Username:</label>
            <input type="text" name="pma_username" required>
            <label>Password:</label>
            <input type="password" name="pma_password" required>
            <input type="submit" value="Log in">
        </form>
        <!-- Default: root / root -->
    </div>
</body>
</html>`, p.config.Applications.PHPMyAdmin.Version, errorMsg)
}

// generateDashboard generates fake dashboard
func (p *PHPMyAdminApp) generateDashboard() string {
	return `<!DOCTYPE html>
<html>
<head><title>phpMyAdmin</title></head>
<body>
    <h1>phpMyAdmin Dashboard</h1>
    <h2>Databases</h2>
    <ul>
        <li>information_schema</li>
        <li>mysql</li>
        <li>performance_schema</li>
        <li>wordpress_db</li>
        <li>customer_data</li>
    </ul>
</body>
</html>`
}
