package http

import (
	"database/sql"
	"fmt"
	"net/http"
	"strings"

	"honey/internal/config"
	"honey/internal/database"
	"honey/internal/logger"
	"honey/internal/models"
	"time"
)

// WordPressApp emulates a WordPress installation
type WordPressApp struct {
	config   *config.HTTPConfig
	logger   logger.Logger
	db       *sql.DB
	detector *VulnerabilityDetector
}

// NewWordPressApp creates a new WordPress honeypot application
func NewWordPressApp(cfg *config.HTTPConfig, log logger.Logger, db *sql.DB, detector *VulnerabilityDetector) *WordPressApp {
	return &WordPressApp{
		config:   cfg,
		logger:   log,
		db:       db,
		detector: detector,
	}
}

// Paths returns the paths this application handles
func (w *WordPressApp) Paths() []string {
	basePath := w.config.Applications.WordPress.Path
	return []string{
		basePath + "/wp-login.php",
		basePath + "/wp-admin",
		basePath + "/wp-admin/",
		basePath + "/wp-content",
		basePath + "/wp-includes",
		basePath + "/xmlrpc.php",
		basePath + "/wp-json",
		basePath + "/wp-config.php",
		basePath + "/readme.html",
		basePath + "/license.txt",
		basePath,
		basePath + "/",
	}
}

// HandleRequest handles WordPress requests
func (w *WordPressApp) HandleRequest(r *http.Request, body string, requestID int) *Response {
	path := r.URL.Path
	basePath := w.config.Applications.WordPress.Path

	// Run vulnerability detection
	w.detector.Analyze(r, r.URL.RawQuery+" "+body, requestID)

	// Handle specific endpoints
	if strings.HasSuffix(path, "/wp-login.php") {
		return w.handleLogin(r, body, requestID)
	}

	if strings.Contains(path, "/wp-admin") {
		return w.handleAdmin(r, requestID)
	}

	if strings.HasSuffix(path, "/xmlrpc.php") {
		return w.handleXMLRPC(r, body, requestID)
	}

	if strings.Contains(path, "/wp-json") {
		return w.handleWPJSON(r, requestID)
	}

	if strings.HasSuffix(path, "/wp-config.php") {
		return w.handleWPConfig(requestID)
	}

	if strings.HasSuffix(path, "/readme.html") {
		return w.handleReadme()
	}

	if strings.HasSuffix(path, "/license.txt") {
		return w.handleLicense()
	}

	// Default WordPress homepage
	if path == basePath || path == basePath+"/" {
		return w.handleHomepage()
	}

	return &Response{
		StatusCode: 404,
		Headers:    map[string][]string{"Content-Type": {"text/html"}},
		Body:       "404 Not Found",
	}
}

// handleLogin handles WordPress login page and attempts
func (w *WordPressApp) handleLogin(r *http.Request, body string, requestID int) *Response {
	if r.Method == "POST" {
		// Parse form data
		username := ""
		password := ""

		// Simple form parsing
		parts := strings.Split(body, "&")
		for _, part := range parts {
			kv := strings.SplitN(part, "=", 2)
			if len(kv) == 2 {
				if kv[0] == "log" {
					username = kv[1]
				}
				if kv[0] == "pwd" {
					password = kv[1]
				}
			}
		}

		w.logger.Infof("[WordPress] Login attempt - Username: %s, Password: %s", username, password)

		// Save credentials
		cred := &models.HTTPCredential{
			RequestID:   requestID,
			Application: "wordpress",
			Username:    username,
			Password:    password,
			Success:     false,
			Timestamp:   time.Now(),
		}

		// Check if credentials match fake users
		for _, user := range w.config.Applications.WordPress.FakeCredentials {
			if username == user.Username && password == user.Password {
				cred.Success = true
				w.logger.Warnf("[WordPress] Successful login with fake credentials: %s/%s", username, password)
				break
			}
		}

		database.SaveHTTPCredential(cred)

		if cred.Success {
			// Redirect to admin
			return &Response{
				StatusCode: 302,
				Headers: map[string][]string{
					"Location":     {w.config.Applications.WordPress.Path + "/wp-admin/"},
					"Content-Type": {"text/html"},
				},
				Body: "Redirecting...",
			}
		}

		// Failed login
		return &Response{
			StatusCode: 200,
			Headers:    map[string][]string{"Content-Type": {"text/html"}},
			Body:       w.generateLoginPage(true),
		}
	}

	// GET request - show login form
	return &Response{
		StatusCode: 200,
		Headers:    map[string][]string{"Content-Type": {"text/html"}},
		Body:       w.generateLoginPage(false),
	}
}

// handleAdmin handles wp-admin requests
func (w *WordPressApp) handleAdmin(r *http.Request, requestID int) *Response {
	// Check for auth cookie (simple check)
	cookie, _ := r.Cookie("wordpress_logged_in")
	if cookie == nil {
		// Redirect to login
		return &Response{
			StatusCode: 302,
			Headers: map[string][]string{
				"Location":     {w.config.Applications.WordPress.Path + "/wp-login.php"},
				"Content-Type": {"text/html"},
			},
			Body: "Redirecting to login...",
		}
	}

	// Show fake admin dashboard
	return &Response{
		StatusCode: 200,
		Headers:    map[string][]string{"Content-Type": {"text/html"}},
		Body:       w.generateAdminDashboard(),
	}
}

// handleXMLRPC handles XML-RPC requests (common brute force target)
func (w *WordPressApp) handleXMLRPC(r *http.Request, body string, requestID int) *Response {
	w.logger.Warnf("[WordPress] XML-RPC request from %s", r.RemoteAddr)

	// Log the XML-RPC method called
	if strings.Contains(body, "wp.getUsersBlogs") || strings.Contains(body, "system.multicall") {
		w.logger.Warnf("[WordPress] Potential XML-RPC brute force detected")
	}

	return &Response{
		StatusCode: 200,
		Headers:    map[string][]string{"Content-Type": {"text/xml"}},
		Body: `<?xml version="1.0" encoding="UTF-8"?>
<methodResponse>
  <params>
    <param>
      <value>
        <array>
          <data>
            <value><string>WordPress XMLRPC API</string></value>
          </data>
        </array>
      </value>
    </param>
  </params>
</methodResponse>`,
	}
}

// handleWPJSON handles WordPress REST API requests
func (w *WordPressApp) handleWPJSON(r *http.Request, requestID int) *Response {
	// User enumeration endpoint
	if strings.Contains(r.URL.Path, "/wp-json/wp/v2/users") {
		w.logger.Warnf("[WordPress] User enumeration attempt")
		return &Response{
			StatusCode: 200,
			Headers:    map[string][]string{"Content-Type": {"application/json"}},
			Body: `[
  {"id":1,"name":"admin","slug":"admin","link":"http://localhost/author/admin/"},
  {"id":2,"name":"editor","slug":"editor","link":"http://localhost/author/editor/"}
]`,
		}
	}

	return &Response{
		StatusCode: 200,
		Headers:    map[string][]string{"Content-Type": {"application/json"}},
		Body:       `{"name":"WordPress Site","description":"Just another WordPress site","url":"http://localhost","home":"http://localhost"}`,
	}
}

// handleWPConfig handles wp-config.php requests
func (w *WordPressApp) handleWPConfig(requestID int) *Response {
	w.logger.Warnf("[WordPress] Attempt to access wp-config.php")

	return &Response{
		StatusCode: 200,
		Headers:    map[string][]string{"Content-Type": {"text/plain"}},
		Body: `<?php
define('DB_NAME', 'wordpress_db');
define('DB_USER', 'wp_user');
define('DB_PASSWORD', 'wp_P@ssw0rd123');
define('DB_HOST', 'localhost');
define('DB_CHARSET', 'utf8mb4');

define('AUTH_KEY',         'put your unique phrase here');
define('SECURE_AUTH_KEY',  'put your unique phrase here');
define('LOGGED_IN_KEY',    'put your unique phrase here');
define('NONCE_KEY',        'put your unique phrase here');

$table_prefix = 'wp_';
define('WP_DEBUG', true);

/* Admin credentials (for development only!)
   Username: admin
   Password: admin123
*/
?>`,
	}
}

// handleReadme handles readme.html (version disclosure)
func (w *WordPressApp) handleReadme() *Response {
	return &Response{
		StatusCode: 200,
		Headers:    map[string][]string{"Content-Type": {"text/html"}},
		Body: `<!DOCTYPE html>
<html>
<head>
	<meta charset="UTF-8">
	<title>WordPress ` + w.config.Applications.WordPress.Version + `</title>
</head>
<body>
	<h1>WordPress</h1>
	<p>Version ` + w.config.Applications.WordPress.Version + `</p>
	<p>Semantic Personal Publishing Platform</p>
</body>
</html>`,
	}
}

// handleLicense handles license.txt
func (w *WordPressApp) handleLicense() *Response {
	return &Response{
		StatusCode: 200,
		Headers:    map[string][]string{"Content-Type": {"text/plain"}},
		Body:       "WordPress - Web publishing software\nVersion " + w.config.Applications.WordPress.Version,
	}
}

// handleHomepage handles WordPress homepage
func (w *WordPressApp) handleHomepage() *Response {
	return &Response{
		StatusCode: 200,
		Headers:    map[string][]string{"Content-Type": {"text/html"}},
		Body:       w.generateHomepage(),
	}
}

// generateLoginPage generates the WordPress login page
func (w *WordPressApp) generateLoginPage(failed bool) string {
	errorMsg := ""
	if failed {
		errorMsg = `<div class="error"><strong>ERROR</strong>: Invalid username or password.</div>`
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en-US">
<head>
	<meta charset="UTF-8">
	<title>Log In &lsaquo; WordPress Site &mdash; WordPress</title>
	<meta name="robots" content="noindex, nofollow">
	<style>
		body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Oxygen-Sans, Ubuntu, Cantarell, "Helvetica Neue", sans-serif; background: #f1f1f1; }
		#login { width: 320px; padding: 8%% 0 0; margin: auto; }
		.login h1 a { background-image: url(data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iODQiIGhlaWdodD0iODQiIHZpZXdCb3g9IjAgMCA4NCA4NCIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj48ZyBmaWxsPSJub25lIiBmaWxsLXJ1bGU9ImV2ZW5vZGQiPjxwYXRoIGZpbGw9IiMwMDczYWEiIGQ9Ik00MiA0MmgyMHYyMEg0MnoiLz48L2c+PC9zdmc+); width: 84px; height: 84px; margin: 0 auto 25px; }
		.login form { margin-top: 20px; margin-left: 0; padding: 26px 24px; font-weight: 400; background: #fff; border: 1px solid #c3c4c7; box-shadow: 0 1px 3px rgb(0 0 0 / 4%%); }
		.login label { font-size: 14px; line-height: 1.5; display: inline-block; margin-bottom: 3px; }
		.login input[type=text], .login input[type=password] { font-size: 24px; line-height: 1.33333333; width: 100%%; border: 1px solid #949494; padding: 3px 5px; margin: 0 6px 16px 0; }
		.login input[type=submit] { float: left; width: auto; height: 32px; line-height: 30px; padding: 0 12px 2px; background: #2271b1; border-color: #2271b1; color: #fff; text-decoration: none; text-shadow: none; }
		.error { margin: 0 0 16px 8px; border-left: 4px solid #d63638; padding: 12px; background-color: #fff; box-shadow: 0 1px 1px 0 rgb(0 0 0 / 10%%); }
	</style>
</head>
<body class="login">
	<div id="login">
		<h1 class="login"><a href="https://wordpress.org/">Powered by WordPress</a></h1>
		%s
		<form name="loginform" id="loginform" action="%s/wp-login.php" method="post">
			<p>
				<label for="user_login">Username or Email Address</label>
				<input type="text" name="log" id="user_login" class="input" value="" size="20" autocapitalize="off" />
			</p>
			<p>
				<label for="user_pass">Password</label>
				<input type="password" name="pwd" id="user_pass" class="input" value="" size="20" />
			</p>
			<p class="submit">
				<input type="submit" name="wp-submit" id="wp-submit" class="button button-primary button-large" value="Log In" />
			</p>
		</form>
		<!-- Debug credentials: admin / admin123 -->
	</div>
</body>
</html>`, errorMsg, w.config.Applications.WordPress.Path)
}

// generateHomepage generates a realistic WordPress homepage
func (w *WordPressApp) generateHomepage() string {
	return `<!DOCTYPE html>
<html lang="en-US">
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1">
	<meta name="generator" content="WordPress ` + w.config.Applications.WordPress.Version + `" />
	<title>TechBlog - Web Development &amp; Security News</title>
	<link rel="profile" href="https://gmpg.org/xfn/11">
	<style>
		* { margin: 0; padding: 0; box-sizing: border-box; }
		body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif; line-height: 1.6; color: #333; background: #f5f5f5; }

		/* Header */
		.site-header { background: #2c3e50; color: white; padding: 20px 0; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
		.site-header .container { max-width: 1200px; margin: 0 auto; padding: 0 20px; display: flex; justify-content: space-between; align-items: center; }
		.site-title { font-size: 28px; font-weight: 700; }
		.site-title a { color: white; text-decoration: none; }
		.site-description { font-size: 14px; color: #ecf0f1; margin-top: 5px; }

		/* Navigation */
		.main-navigation { margin-top: 10px; }
		.main-navigation ul { list-style: none; display: flex; gap: 30px; }
		.main-navigation a { color: #ecf0f1; text-decoration: none; font-size: 16px; transition: color 0.3s; }
		.main-navigation a:hover { color: #3498db; }

		/* Search */
		.header-search { display: flex; align-items: center; }
		.search-form { display: flex; }
		.search-form input[type="search"] { padding: 8px 12px; border: none; border-radius: 4px 0 0 4px; width: 200px; }
		.search-form button { padding: 8px 15px; background: #3498db; color: white; border: none; border-radius: 0 4px 4px 0; cursor: pointer; }

		/* Main Content */
		.site-content { max-width: 1200px; margin: 40px auto; padding: 0 20px; display: grid; grid-template-columns: 1fr 350px; gap: 40px; }

		/* Posts */
		.posts-list { display: flex; flex-direction: column; gap: 30px; }
		.post { background: white; border-radius: 8px; overflow: hidden; box-shadow: 0 2px 8px rgba(0,0,0,0.1); transition: transform 0.3s; }
		.post:hover { transform: translateY(-2px); box-shadow: 0 4px 12px rgba(0,0,0,0.15); }
		.post-thumbnail { width: 100%; height: 300px; background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); display: flex; align-items: center; justify-content: center; color: white; font-size: 24px; }
		.post-content { padding: 30px; }
		.post-title { font-size: 28px; margin-bottom: 10px; }
		.post-title a { color: #2c3e50; text-decoration: none; }
		.post-title a:hover { color: #3498db; }
		.post-meta { color: #7f8c8d; font-size: 14px; margin-bottom: 15px; }
		.post-meta a { color: #3498db; text-decoration: none; }
		.post-excerpt { color: #555; line-height: 1.8; margin-bottom: 15px; }
		.read-more { display: inline-block; color: #3498db; text-decoration: none; font-weight: 600; }
		.read-more:hover { text-decoration: underline; }

		/* Sidebar */
		.sidebar { display: flex; flex-direction: column; gap: 30px; }
		.widget { background: white; border-radius: 8px; padding: 25px; box-shadow: 0 2px 8px rgba(0,0,0,0.1); }
		.widget-title { font-size: 20px; margin-bottom: 15px; color: #2c3e50; border-bottom: 2px solid #3498db; padding-bottom: 10px; }
		.widget ul { list-style: none; }
		.widget li { padding: 8px 0; border-bottom: 1px solid #ecf0f1; }
		.widget li:last-child { border-bottom: none; }
		.widget a { color: #555; text-decoration: none; transition: color 0.3s; }
		.widget a:hover { color: #3498db; }

		/* Footer */
		.site-footer { background: #2c3e50; color: #ecf0f1; padding: 40px 0 20px; margin-top: 60px; }
		.site-footer .container { max-width: 1200px; margin: 0 auto; padding: 0 20px; text-align: center; }
		.footer-text { font-size: 14px; }
		.footer-text a { color: #3498db; text-decoration: none; }

		/* Tags */
		.tag { display: inline-block; background: #ecf0f1; padding: 4px 12px; border-radius: 3px; font-size: 12px; color: #555; margin-right: 5px; }
	</style>
</head>
<body class="home blog">

<!-- Header -->
<header class="site-header">
	<div class="container">
		<div>
			<h1 class="site-title"><a href="/wordpress/">TechBlog</a></h1>
			<p class="site-description">Web Development, Security & Tech News</p>
		</div>
		<div class="header-search">
			<form class="search-form" action="/wordpress/" method="get">
				<input type="search" name="s" placeholder="Search..." />
				<button type="submit">Search</button>
			</form>
		</div>
	</div>
	<div class="container">
		<nav class="main-navigation">
			<ul>
				<li><a href="/wordpress/">Home</a></li>
				<li><a href="/wordpress/category/tutorials">Tutorials</a></li>
				<li><a href="/wordpress/category/news">News</a></li>
				<li><a href="/wordpress/about">About</a></li>
				<li><a href="/wordpress/contact">Contact</a></li>
				<li><a href="/wordpress/wp-admin">Admin</a></li>
			</ul>
		</nav>
	</div>
</header>

<!-- Main Content -->
<div class="site-content">
	<!-- Posts -->
	<main class="posts-list">
		<article class="post">
			<div class="post-thumbnail">📱 Featured Image</div>
			<div class="post-content">
				<h2 class="post-title"><a href="/wordpress/2025/01/15/building-secure-rest-apis/">Building Secure REST APIs with Node.js</a></h2>
				<div class="post-meta">
					Posted on <a href="#">January 15, 2025</a> by <a href="#">admin</a> |
					<a href="#">Web Development</a>, <a href="#">Security</a>
				</div>
				<div class="post-excerpt">
					<p>Learn how to build production-ready REST APIs with proper authentication, input validation, and security best practices. This comprehensive guide covers JWT authentication, rate limiting, SQL injection prevention, and more...</p>
				</div>
				<a href="/wordpress/2025/01/15/building-secure-rest-apis/" class="read-more">Continue Reading →</a>
			</div>
		</article>

		<article class="post">
			<div class="post-thumbnail" style="background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);">🔒 Featured Image</div>
			<div class="post-content">
				<h2 class="post-title"><a href="/wordpress/2025/01/12/common-web-vulnerabilities/">Top 10 Web Vulnerabilities in 2025</a></h2>
				<div class="post-meta">
					Posted on <a href="#">January 12, 2025</a> by <a href="#">editor</a> |
					<a href="#">Security</a>
				</div>
				<div class="post-excerpt">
					<p>OWASP has released their updated list of the most critical web application security risks. From SQL injection to broken authentication, learn about these vulnerabilities and how to protect your applications...</p>
				</div>
				<a href="/wordpress/2025/01/12/common-web-vulnerabilities/" class="read-more">Continue Reading →</a>
			</div>
		</article>

		<article class="post">
			<div class="post-thumbnail" style="background: linear-gradient(135deg, #4facfe 0%, #00f2fe 100%);">⚛️ Featured Image</div>
			<div class="post-content">
				<h2 class="post-title"><a href="/wordpress/2025/01/08/react-performance-tips/">React Performance Optimization Tips</a></h2>
				<div class="post-meta">
					Posted on <a href="#">January 8, 2025</a> by <a href="#">admin</a> |
					<a href="#">React</a>, <a href="#">Performance</a>
				</div>
				<div class="post-excerpt">
					<p>Struggling with slow React applications? Discover proven techniques to optimize your React apps including memoization, lazy loading, code splitting, and virtual scrolling. Make your apps blazing fast...</p>
				</div>
				<a href="/wordpress/2025/01/08/react-performance-tips/" class="read-more">Continue Reading →</a>
			</div>
		</article>
	</main>

	<!-- Sidebar -->
	<aside class="sidebar">
		<div class="widget">
			<h3 class="widget-title">About This Blog</h3>
			<p>Welcome to TechBlog! We share the latest tutorials, news, and insights about web development, cybersecurity, and emerging technologies.</p>
		</div>

		<div class="widget">
			<h3 class="widget-title">Recent Posts</h3>
			<ul>
				<li><a href="#">Building Secure REST APIs with Node.js</a></li>
				<li><a href="#">Top 10 Web Vulnerabilities in 2025</a></li>
				<li><a href="#">React Performance Optimization Tips</a></li>
				<li><a href="#">Getting Started with Docker Containers</a></li>
				<li><a href="#">Understanding JWT Authentication</a></li>
			</ul>
		</div>

		<div class="widget">
			<h3 class="widget-title">Categories</h3>
			<ul>
				<li><a href="#">Web Development (15)</a></li>
				<li><a href="#">Security (12)</a></li>
				<li><a href="#">Tutorials (23)</a></li>
				<li><a href="#">DevOps (8)</a></li>
				<li><a href="#">JavaScript (18)</a></li>
			</ul>
		</div>

		<div class="widget">
			<h3 class="widget-title">Tags</h3>
			<div>
				<span class="tag">React</span>
				<span class="tag">Node.js</span>
				<span class="tag">Security</span>
				<span class="tag">API</span>
				<span class="tag">Docker</span>
				<span class="tag">Python</span>
				<span class="tag">JavaScript</span>
				<span class="tag">CSS</span>
			</div>
		</div>

		<div class="widget">
			<h3 class="widget-title">Archives</h3>
			<ul>
				<li><a href="#">January 2025 (3)</a></li>
				<li><a href="#">December 2024 (5)</a></li>
				<li><a href="#">November 2024 (7)</a></li>
				<li><a href="#">October 2024 (6)</a></li>
			</ul>
		</div>
	</aside>
</div>

<!-- Footer -->
<footer class="site-footer">
	<div class="container">
		<p class="footer-text">© 2025 TechBlog. All rights reserved. | Powered by <a href="https://wordpress.org">WordPress ` + w.config.Applications.WordPress.Version + `</a></p>
		<p style="margin-top: 10px; font-size: 12px; color: #95a5a6;">
			<a href="/wordpress/privacy-policy">Privacy Policy</a> |
			<a href="/wordpress/terms">Terms of Service</a> |
			<a href="/wordpress/wp-login.php">Login</a>
		</p>
	</div>
</footer>

<!-- WordPress Meta -->
<link rel='dns-prefetch' href='//s.w.org' />
<script type='text/javascript' src='/wordpress/wp-includes/js/jquery/jquery.min.js?ver=3.6.0'></script>
<script type='text/javascript' src='/wordpress/wp-includes/js/jquery/jquery-migrate.min.js?ver=3.3.2'></script>
<!-- WP Debug: DB credentials in wp-config.php -->
<!-- WP_DEBUG is enabled. Database: wordpress_db -->

</body>
</html>`
}

// generateAdminDashboard generates a fake WordPress admin dashboard
func (w *WordPressApp) generateAdminDashboard() string {
	return `<!DOCTYPE html>
<html lang="en-US">
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1">
	<title>Dashboard ‹ TechBlog — WordPress</title>
	<style>
		* { margin: 0; padding: 0; box-sizing: border-box; }
		body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Oxygen-Sans, Ubuntu, Cantarell, "Helvetica Neue", sans-serif; background: #f0f0f1; color: #3c434a; }

		#wpadminbar { background: #1d2327; color: #fff; height: 32px; padding: 0 12px; display: flex; align-items: center; font-size: 13px; }
		#wpadminbar a { color: #fff; text-decoration: none; margin: 0 10px; }

		#adminmenuback, #adminmenuwrap { width: 160px; position: fixed; top: 32px; bottom: 0; left: 0; background: #1d2327; }
		#adminmenu { margin: 0; padding: 0; list-style: none; }
		#adminmenu li { border-bottom: 1px solid #2c3338; }
		#adminmenu a { display: block; padding: 12px 15px; color: #c3c4c7; text-decoration: none; font-size: 14px; }
		#adminmenu a:hover { background: #2c3338; color: #72aee6; }

		#wpcontent { margin-left: 160px; padding: 0; }
		#wpbody { padding: 20px 40px 0 20px; }
		#wpbody-content { padding: 20px; }

		.wrap { max-width: 1200px; }
		h1 { font-size: 23px; font-weight: 400; margin: 0 0 20px; padding: 9px 0 4px; line-height: 1.3; }

		.welcome-panel { background: #fff; border: 1px solid #c3c4c7; box-shadow: 0 1px 1px rgb(0 0 0 / 4%); padding: 23px 10px 12px; margin: 20px 0; }
		.welcome-panel-content { max-width: 1200px; margin: 0 auto; }
		.welcome-panel h2 { margin: 0; font-size: 21px; font-weight: 400; line-height: 1.2; }
		.welcome-panel p { font-size: 13px; }

		#dashboard-widgets { display: grid; grid-template-columns: 1fr 1fr; gap: 20px; margin-top: 20px; }
		.postbox { background: #fff; border: 1px solid #c3c4c7; box-shadow: 0 1px 1px rgb(0 0 0 / 4%); }
		.postbox-header { border-bottom: 1px solid #c3c4c7; padding: 12px; }
		.postbox-header h2 { font-size: 14px; font-weight: 600; margin: 0; }
		.inside { padding: 12px; }

		.activity-block { margin-bottom: 15px; padding-bottom: 15px; border-bottom: 1px solid #dcdcde; }
		.activity-block:last-child { border-bottom: none; }
		.activity-block h3 { font-size: 14px; margin: 0 0 5px; }
		.subsubsub { color: #646970; font-size: 13px; }

		ul.stats { list-style: none; }
		ul.stats li { padding: 8px 0; border-bottom: 1px solid #f0f0f1; display: flex; justify-content: space-between; }
		.stats-number { font-weight: 600; color: #2271b1; }
	</style>
</head>
<body class="wp-admin wp-core-ui">

<!-- Admin Bar -->
<div id="wpadminbar">
	<a href="/wordpress/">Visit Site</a>
	<a href="#">TechBlog</a>
	<span style="margin-left: auto;">Hi, <strong>admin</strong> | <a href="/wordpress/wp-login.php?action=logout">Log Out</a></span>
</div>

<!-- Admin Menu -->
<div id="adminmenuwrap">
	<ul id="adminmenu">
		<li><a href="/wordpress/wp-admin/">Dashboard</a></li>
		<li><a href="/wordpress/wp-admin/post-new.php">Posts</a></li>
		<li><a href="/wordpress/wp-admin/edit.php?post_type=page">Pages</a></li>
		<li><a href="/wordpress/wp-admin/edit-comments.php">Comments</a></li>
		<li><a href="/wordpress/wp-admin/themes.php">Appearance</a></li>
		<li><a href="/wordpress/wp-admin/plugins.php">Plugins</a></li>
		<li><a href="/wordpress/wp-admin/users.php">Users</a></li>
		<li><a href="/wordpress/wp-admin/tools.php">Tools</a></li>
		<li><a href="/wordpress/wp-admin/options-general.php">Settings</a></li>
	</ul>
</div>

<!-- Main Content -->
<div id="wpcontent">
	<div id="wpbody">
		<div id="wpbody-content">
			<div class="wrap">
				<h1>Dashboard</h1>

				<div class="welcome-panel">
					<div class="welcome-panel-content">
						<h2>Welcome to WordPress!</h2>
						<p class="about-description">We've assembled some links to get you started:</p>
						<div style="margin-top: 15px;">
							<a href="/wordpress/wp-admin/post-new.php" style="margin-right: 15px; color: #2271b1; text-decoration: none;">Write your first blog post</a>
							<a href="/wordpress/wp-admin/customize.php" style="margin-right: 15px; color: #2271b1; text-decoration: none;">Customize your site</a>
							<a href="/wordpress/wp-admin/options-general.php" style="color: #2271b1; text-decoration: none;">Manage settings</a>
						</div>
					</div>
				</div>

				<div id="dashboard-widgets">
					<div class="postbox">
						<div class="postbox-header">
							<h2>At a Glance</h2>
						</div>
						<div class="inside">
							<ul class="stats">
								<li><span>Posts</span> <span class="stats-number">15</span></li>
								<li><span>Pages</span> <span class="stats-number">8</span></li>
								<li><span>Comments</span> <span class="stats-number">42</span></li>
								<li><span>Categories</span> <span class="stats-number">6</span></li>
							</ul>
							<p style="margin-top: 12px; color: #646970; font-size: 13px;">
								Running WordPress ` + w.config.Applications.WordPress.Version + `<br>
								Active Theme: TwentyTwentyFour<br>
								PHP Version: 8.1.2
							</p>
						</div>
					</div>

					<div class="postbox">
						<div class="postbox-header">
							<h2>Activity</h2>
						</div>
						<div class="inside">
							<div class="activity-block">
								<h3>Recently Published</h3>
								<ul style="list-style: none; font-size: 13px;">
									<li style="padding: 5px 0;">📝 Building Secure REST APIs with Node.js</li>
									<li style="padding: 5px 0;">📝 Top 10 Web Vulnerabilities in 2025</li>
									<li style="padding: 5px 0;">📝 React Performance Optimization Tips</li>
								</ul>
							</div>
							<div class="activity-block">
								<h3>Recent Comments</h3>
								<ul style="list-style: none; font-size: 13px;">
									<li style="padding: 5px 0;">💬 John on "Building Secure REST APIs"</li>
									<li style="padding: 5px 0;">💬 Sarah on "Web Vulnerabilities"</li>
								</ul>
							</div>
						</div>
					</div>

					<div class="postbox">
						<div class="postbox-header">
							<h2>Quick Draft</h2>
						</div>
						<div class="inside">
							<form>
								<input type="text" placeholder="Title" style="width: 100%; padding: 8px; margin-bottom: 10px; border: 1px solid #8c8f94; border-radius: 3px;" />
								<textarea placeholder="What's on your mind?" style="width: 100%; height: 100px; padding: 8px; border: 1px solid #8c8f94; border-radius: 3px;"></textarea>
								<button type="submit" style="margin-top: 10px; padding: 8px 15px; background: #2271b1; color: #fff; border: none; border-radius: 3px; cursor: pointer;">Save Draft</button>
							</form>
						</div>
					</div>

					<div class="postbox">
						<div class="postbox-header">
							<h2>WordPress News</h2>
						</div>
						<div class="inside">
							<div style="font-size: 13px;">
								<div style="margin-bottom: 15px;">
									<strong>WordPress 6.5 Beta Now Available</strong>
									<p style="color: #646970; margin: 5px 0;">The first beta for WordPress 6.5 is now available for testing...</p>
								</div>
								<div style="margin-bottom: 15px;">
									<strong>Security Update Released</strong>
									<p style="color: #646970; margin: 5px 0;">WordPress 5.8.2 security and maintenance release...</p>
								</div>
							</div>
						</div>
					</div>
				</div>
			</div>
		</div>
	</div>
</div>

</body>
</html>`
}
