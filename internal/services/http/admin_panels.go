package http

import (
	"database/sql"
	"net/http"
	"strings"
	"time"

	"honey/internal/config"
	"honey/internal/database"
	"honey/internal/logger"
	"honey/internal/models"
)

// AdminPanelApp emulates various admin panels
type AdminPanelApp struct {
	config   *config.HTTPConfig
	logger   logger.Logger
	db       *sql.DB
	detector *VulnerabilityDetector
}

// NewAdminPanelApp creates a new admin panel honeypot
func NewAdminPanelApp(cfg *config.HTTPConfig, log logger.Logger, db *sql.DB, detector *VulnerabilityDetector) *AdminPanelApp {
	return &AdminPanelApp{
		config:   cfg,
		logger:   log,
		db:       db,
		detector: detector,
	}
}

// Paths returns handled paths
func (a *AdminPanelApp) Paths() []string {
	return []string{
		"/admin",
		"/admin/",
		"/admin/login",
		"/administrator",
		"/administrator/",
		"/admin.php",
		"/login.php",
		"/manager/html",
		"/cpanel",
		"/cpanel/",
	}
}

// HandleRequest handles admin panel requests
func (a *AdminPanelApp) HandleRequest(r *http.Request, body string, requestID int) *Response {
	// Run vulnerability detection
	a.detector.Analyze(r, r.URL.RawQuery+" "+body, requestID)

	if r.Method == "POST" {
		return a.handleLogin(r, body, requestID)
	}

	return &Response{
		StatusCode: 200,
		Headers:    map[string][]string{"Content-Type": {"text/html"}},
		Body:       a.generateLoginPage(r.URL.Path),
	}
}

// handleLogin handles login attempts
func (a *AdminPanelApp) handleLogin(r *http.Request, body string, requestID int) *Response {
	username := ""
	password := ""

	parts := strings.Split(body, "&")
	for _, part := range parts {
		kv := strings.SplitN(part, "=", 2)
		if len(kv) == 2 {
			if strings.Contains(kv[0], "user") || kv[0] == "login" {
				username = kv[1]
			}
			if strings.Contains(kv[0], "pass") || kv[0] == "pwd" {
				password = kv[1]
			}
		}
	}

	a.logger.Infof("[Admin Panel] Login attempt - Username: %s, Password: %s", username, password)

	cred := &models.HTTPCredential{
		RequestID:   requestID,
		Application: "admin_panel",
		Username:    username,
		Password:    password,
		Success:     (username == "admin" && password == "admin"),
		Timestamp:   time.Now(),
	}

	database.SaveHTTPCredential(cred)

	return &Response{
		StatusCode: 200,
		Headers:    map[string][]string{"Content-Type": {"text/html"}},
		Body:       a.generateLoginPage(r.URL.Path),
	}
}

// generateLoginPage generates an admin login page
func (a *AdminPanelApp) generateLoginPage(path string) string {
	return `<!DOCTYPE html>
<html>
<head>
    <title>Admin Login</title>
    <style>
        body { font-family: Arial; background: #2c3e50; display: flex; justify-content: center; align-items: center; height: 100vh; margin: 0; }
        .login-box { background: white; padding: 40px; border-radius: 5px; box-shadow: 0 0 20px rgba(0,0,0,0.3); width: 300px; }
        h2 { text-align: center; color: #2c3e50; margin-bottom: 30px; }
        input { width: 100%; padding: 12px; margin: 8px 0; box-sizing: border-box; border: 1px solid #ddd; border-radius: 3px; }
        button { width: 100%; padding: 12px; background: #3498db; color: white; border: none; border-radius: 3px; cursor: pointer; font-size: 16px; }
        button:hover { background: #2980b9; }
    </style>
</head>
<body>
    <div class="login-box">
        <h2>Administrator</h2>
        <form method="post">
            <input type="text" name="username" placeholder="Username" required>
            <input type="password" name="password" placeholder="Password" required>
            <button type="submit">Login</button>
        </form>
    </div>
</body>
</html>`
}
