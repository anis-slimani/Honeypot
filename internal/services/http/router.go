package http

import (
	"database/sql"
	"net/http"
	"strings"

	"honey/internal/config"
	"honey/internal/logger"
)

// Router handles virtual host routing to different honeypot applications
type Router struct {
	config       *config.HTTPConfig
	logger       logger.Logger
	db           *sql.DB
	applications map[string]Application
	detector     *VulnerabilityDetector
}

// Response represents an HTTP response
type Response struct {
	StatusCode int
	Headers    map[string][]string
	Body       string
}

// Application interface for honeypot applications
type Application interface {
	HandleRequest(r *http.Request, body string, requestID int) *Response
	Paths() []string
}

// NewRouter creates a new router instance
func NewRouter(cfg *config.HTTPConfig, log logger.Logger, db *sql.DB) *Router {
	router := &Router{
		config:       cfg,
		logger:       log,
		db:           db,
		applications: make(map[string]Application),
		detector:     NewVulnerabilityDetector(log, db),
	}

	// Register applications
	router.registerApplications()

	return router
}

// SetAlertManager définit le gestionnaire d'alertes pour le router et le detector
func (r *Router) SetAlertManager(am AlertManager) {
	if r.detector != nil {
		r.detector.SetAlertManager(am)
	}
}

// AlertManager interface pour envoyer des alertes
type AlertManager interface {
	SendHTTPAlert(alertType, severity, message, remoteAddr, details string)
}

// registerApplications registers all honeypot applications
func (r *Router) registerApplications() {
	// WordPress
	if r.config.Applications.WordPress.Enabled {
		wp := NewWordPressApp(r.config, r.logger, r.db, r.detector)
		for _, path := range wp.Paths() {
			r.applications[path] = wp
		}
		r.logger.Info("WordPress honeypot application registered")
	}

	// phpMyAdmin
	if r.config.Applications.PHPMyAdmin.Enabled {
		pma := NewPHPMyAdminApp(r.config, r.logger, r.db, r.detector)
		for _, path := range pma.Paths() {
			r.applications[path] = pma
		}
		r.logger.Info("phpMyAdmin honeypot application registered")
	}

	// File Upload
	if r.config.Applications.Upload.Enabled {
		upload := NewUploadApp(r.config, r.logger, r.db, r.detector)
		for _, path := range upload.Paths() {
			r.applications[path] = upload
		}
		r.logger.Info("File Upload honeypot application registered")
	}

	// Admin Panels
	if r.config.Applications.AdminPanels.Enabled {
		admin := NewAdminPanelApp(r.config, r.logger, r.db, r.detector)
		for _, path := range admin.Paths() {
			r.applications[path] = admin
		}
		r.logger.Info("Admin Panel honeypot applications registered")
	}

	// Generic handler for unmatched paths
	r.logger.Info("Router initialized with all applications")
}

// Route routes the request to the appropriate application
func (r *Router) Route(req *http.Request, body string, requestID int) *Response {
	path := req.URL.Path

	// Try exact match first
	if app, exists := r.applications[path]; exists {
		return app.HandleRequest(req, body, requestID)
	}

	// Try prefix match
	for appPath, app := range r.applications {
		if strings.HasPrefix(path, appPath) {
			return app.HandleRequest(req, body, requestID)
		}
	}

	// Check for common attack paths
	attackResponse := r.handleCommonAttackPaths(req, body, requestID)
	if attackResponse != nil {
		return attackResponse
	}

	// Default 404 response with honeytokens
	return r.default404Response(req, requestID)
}

// handleCommonAttackPaths handles common attack paths
func (r *Router) handleCommonAttackPaths(req *http.Request, body string, requestID int) *Response {
	path := req.URL.Path

	// Exposed configuration files
	configFiles := map[string]string{
		"/.env":             r.generateFakeEnvFile(),
		"/.git/config":      r.generateFakeGitConfig(),
		"/.git/HEAD":        "ref: refs/heads/main\n",
		"/config.php":       r.generateFakeConfigPHP(),
		"/web.config":       r.generateFakeWebConfig(),
		"/.htaccess":        r.generateFakeHtaccess(),
		"/database.yml":     r.generateFakeDatabaseYML(),
		"/.aws/credentials": r.generateFakeAWSCredentials(),
		"/backup.sql":       r.generateFakeBackupSQL(),
	}

	if content, exists := configFiles[path]; exists {
		r.logger.Warnf("[ATTACK] Attempt to access sensitive file: %s", path)
		return &Response{
			StatusCode: 200,
			Headers:    map[string][]string{"Content-Type": {"text/plain"}},
			Body:       content,
		}
	}

	// Detect path traversal
	if strings.Contains(path, "..") || strings.Contains(req.URL.RawQuery, "..") {
		r.detector.DetectPathTraversal(req, body, requestID)
		return &Response{
			StatusCode: 200,
			Headers:    map[string][]string{"Content-Type": {"text/plain"}},
			Body:       r.generateFakePasswdFile(),
		}
	}

	return nil
}

// default404Response returns a default 404 response with honeytokens
func (r *Router) default404Response(req *http.Request, requestID int) *Response {
	// Run vulnerability detection on all requests
	r.detector.Analyze(req, req.URL.RawQuery+" "+req.URL.Path, requestID)

	return &Response{
		StatusCode: 404,
		Headers:    map[string][]string{"Content-Type": {"text/html"}},
		Body: `<!DOCTYPE html>
<html>
<head>
    <title>404 Not Found</title>
</head>
<body>
    <h1>Not Found</h1>
    <p>The requested URL was not found on this server.</p>
    <!-- Debug mode enabled. API Key: sk_live_51HYjK2L3mN4o5P6q7R8s9T0u -->
    <!-- Database: mysql://admin:P@ssw0rd123@localhost:3306/production_db -->
    <hr>
    <address>Apache/2.4.41 (Ubuntu) Server at localhost Port 80</address>
</body>
</html>`,
	}
}

// generateFakeEnvFile generates a fake .env file with honeytokens
func (r *Router) generateFakeEnvFile() string {
	return `APP_NAME=ProductionApp
APP_ENV=production
APP_KEY=base64:abcdefghijklmnopqrstuvwxyz1234567890ABCDEF==
APP_DEBUG=true
APP_URL=http://localhost

DB_CONNECTION=mysql
DB_HOST=127.0.0.1
DB_PORT=3306
DB_DATABASE=production_db
DB_USERNAME=admin
DB_PASSWORD=P@ssw0rd123

REDIS_HOST=127.0.0.1
REDIS_PASSWORD=redis_secret_pass
REDIS_PORT=6379

AWS_ACCESS_KEY_ID=AKIAIOSFODNN7EXAMPLE
AWS_SECRET_ACCESS_KEY=wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY
AWS_DEFAULT_REGION=us-east-1

STRIPE_KEY=sk_live_51HYjK2L3mN4o5P6q7R8s9T0u
STRIPE_SECRET=sk_live_abcdefghijklmnopqrstuvwxyz123456

MAIL_MAILER=smtp
MAIL_HOST=smtp.gmail.com
MAIL_PORT=587
MAIL_USERNAME=admin@company.com
MAIL_PASSWORD=email_password_123
`
}

// generateFakeGitConfig generates a fake git config
func (r *Router) generateFakeGitConfig() string {
	return `[core]
	repositoryformatversion = 0
	filemode = true
	bare = false
[remote "origin"]
	url = https://github.com/company/production-app.git
	fetch = +refs/heads/*:refs/remotes/origin/*
[branch "main"]
	remote = origin
	merge = refs/heads/main
[user]
	name = Admin User
	email = admin@company.com
`
}

// generateFakeConfigPHP generates a fake PHP config file
func (r *Router) generateFakeConfigPHP() string {
	return `<?php
define('DB_HOST', 'localhost');
define('DB_USER', 'root');
define('DB_PASSWORD', 'root123');
define('DB_NAME', 'production_db');

define('API_KEY', 'sk_live_51HYjK2L3mN4o5P6q7R8s9T0u');
define('SECRET_KEY', 'super_secret_key_12345');

$admin_username = 'admin';
$admin_password = 'admin123';
?>
`
}

// generateFakeWebConfig generates a fake web.config
func (r *Router) generateFakeWebConfig() string {
	return `<?xml version="1.0" encoding="UTF-8"?>
<configuration>
    <connectionStrings>
        <add name="DefaultConnection"
             connectionString="Server=localhost;Database=ProductionDB;User Id=sa;Password=SqlP@ssw0rd123;"
             providerName="System.Data.SqlClient" />
    </connectionStrings>
    <appSettings>
        <add key="AdminUsername" value="administrator" />
        <add key="AdminPassword" value="Admin@2024" />
        <add key="ApiKey" value="sk_live_51HYjK2L3mN4o5P6q7R8s9T0u" />
    </appSettings>
</configuration>
`
}

// generateFakeHtaccess generates a fake .htaccess
func (r *Router) generateFakeHtaccess() string {
	return `RewriteEngine On
RewriteCond %{REQUEST_FILENAME} !-f
RewriteCond %{REQUEST_FILENAME} !-d
RewriteRule ^(.*)$ index.php/$1 [L]

# Protect sensitive files
<FilesMatch "^\.">
    Order allow,deny
    Deny from all
</FilesMatch>

# Admin credentials (remove in production!)
# Username: admin
# Password: admin123
`
}

// generateFakeDatabaseYML generates a fake database.yml
func (r *Router) generateFakeDatabaseYML() string {
	return `production:
  adapter: postgresql
  encoding: unicode
  database: production_db
  pool: 5
  username: postgres
  password: postgres_pass_123
  host: localhost
  port: 5432

development:
  adapter: postgresql
  encoding: unicode
  database: dev_db
  pool: 5
  username: devuser
  password: devpass123
  host: localhost
`
}

// generateFakeAWSCredentials generates fake AWS credentials
func (r *Router) generateFakeAWSCredentials() string {
	return `[default]
aws_access_key_id = AKIAIOSFODNN7EXAMPLE
aws_secret_access_key = wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY
region = us-east-1

[production]
aws_access_key_id = AKIAI44QH8DHBEXAMPLE
aws_secret_access_key = je7MtGbClwBF/2Zp9Utk/h3yCo8nvbEXAMPLEKEY
region = us-west-2
`
}

// generateFakeBackupSQL generates a fake SQL backup
func (r *Router) generateFakeBackupSQL() string {
	return `-- MySQL dump 10.13  Distrib 8.0.32
-- Server version: 8.0.32-Ubuntu

CREATE DATABASE IF NOT EXISTS production_db;
USE production_db;

CREATE TABLE users (
  id INT PRIMARY KEY AUTO_INCREMENT,
  username VARCHAR(255) NOT NULL,
  password VARCHAR(255) NOT NULL,
  email VARCHAR(255),
  role VARCHAR(50)
);

INSERT INTO users VALUES
(1, 'admin', '$2y$10$abcdefghijklmnopqrstuvwxyz1234567890', 'admin@company.com', 'administrator'),
(2, 'john_doe', '$2y$10$1234567890abcdefghijklmnopqrstuvwxyz', 'john@company.com', 'user');

CREATE TABLE api_keys (
  id INT PRIMARY KEY AUTO_INCREMENT,
  key_value VARCHAR(255),
  service VARCHAR(100)
);

INSERT INTO api_keys VALUES
(1, 'sk_live_51HYjK2L3mN4o5P6q7R8s9T0u', 'stripe'),
(2, 'AIzaSyAbCdEfGhIjKlMnOpQrStUvWxYz1234567', 'google_maps');
`
}

// generateFakePasswdFile generates a fake /etc/passwd file
func (r *Router) generateFakePasswdFile() string {
	return `root:x:0:0:root:/root:/bin/bash
daemon:x:1:1:daemon:/usr/sbin:/usr/sbin/nologin
bin:x:2:2:bin:/bin:/usr/sbin/nologin
sys:x:3:3:sys:/dev:/usr/sbin/nologin
www-data:x:33:33:www-data:/var/www:/usr/sbin/nologin
admin:x:1000:1000:Admin User:/home/admin:/bin/bash
mysql:x:999:999::/var/lib/mysql:/bin/false
`
}
