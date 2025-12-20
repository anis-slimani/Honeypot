# 🎯 Honeypot Dashboard Guide

## Access the Dashboard

**URL:** http://localhost:8080

The dashboard automatically refreshes every **10 seconds** to show real-time data.

---

## Dashboard Tabs

### 1. 📊 **Overview Tab**
Shows combined statistics from both SSH and HTTP honeypots:

**Stats Cards:**
- 🔐 SSH Connections - Total SSH connection attempts
- ✅ SSH Success - Successful SSH logins
- 🌐 HTTP Requests - Total HTTP requests received
- ⚔️ HTTP Attacks - Detected HTTP attacks
- 📁 File Uploads - Total files uploaded
- ☠️ Malware Detected - Malicious files identified
- 👥 Unique Attackers - Distinct IP addresses
- ⚡ Total Commands - SSH commands executed

**Charts:**
- Top Attack Types (SQL injection, XSS, etc.)
- Top Countries (geographic attack sources)

---

### 2. 🔐 **SSH Honeypot Tab**

**Recent SSH Connections:**
- IP Address
- Username attempted
- Password attempted
- Success/Failure status
- Connection timestamp
- Session duration

**SSH Commands Executed:**
- IP Address
- Username
- Command text
- Execution timestamp

---

### 3. 🌐 **HTTP Honeypot Tab**

**Statistics:**
- Total HTTP Requests
- Attacks Detected
- Files Uploaded
- Malware Files

**Recent HTTP Requests:**
- Client IP
- HTTP Method (GET, POST, etc.)
- Request Path
- User-Agent string
- Response Status Code
- Timestamp

**Top Requested Paths:**
- Most frequently accessed URLs
- Request counts per path

**Detected Scanners:**
- IP Address
- Scanner Type (sqlmap, Nikto, Nmap, etc.)
- Version
- Total Requests
- Requests Per Second
- Detection timestamp
- Last seen timestamp

---

### 4. ⚔️ **Attacks Tab**

**HTTP Attacks Detected:**
- Attack Type (SQLi, XSS, Command Injection, Path Traversal, etc.)
- Severity Level (CRITICAL, HIGH, MEDIUM, LOW)
- Attack Payload
- Matched Pattern
- Timestamp

**Severity Color Coding:**
- 🔴 **CRITICAL** - Red (destructive attacks)
- 🟠 **HIGH** - Orange (exploitation attempts)
- 🟡 **MEDIUM** - Yellow (reconnaissance)
- 🔵 **LOW** - Blue (general suspicious activity)

**Security Alerts:**
- Alert Type
- Severity
- Alert Message
- Source IP
- Timestamp

---

### 5. 📁 **Uploads Tab**

**Uploaded Files:**
- Original Filename
- File Size (KB)
- Content Type
- MD5 Hash
- Malware Status (☠️ MALICIOUS / ✓ Safe)
- Upload Timestamp

**File Analysis:**
- Automatic malware detection
- Hash calculation for forensics
- Quarantine in `/root/uploads/`

---

### 6. 🔑 **Credentials Tab**

**SSH Credentials:**
- IP Address
- Username
- Password
- Success/Failure
- Timestamp
- Duration

**HTTP Credentials:**
- Application (WordPress, phpMyAdmin, Admin Panel)
- Username
- Password
- Success/Failure
- Timestamp

---

## API Endpoints

All data is available via REST API:

### SSH Honeypot APIs
```
GET /api/statistics          # SSH statistics
GET /api/connections?limit=N # Recent SSH connections
GET /api/commands?limit=N    # Executed commands
GET /api/alerts?limit=N      # Security alerts
```

### HTTP Honeypot APIs
```
GET /api/http/statistics     # HTTP statistics
GET /api/http/requests?limit=N      # HTTP requests
GET /api/http/attacks?limit=N       # Detected attacks
GET /api/http/uploads?limit=N       # Uploaded files
GET /api/http/credentials?limit=N   # Captured credentials
GET /api/http/scanners?limit=N      # Detected scanners
```

---

## Example API Queries

### Get HTTP Statistics
```bash
curl http://localhost:8080/api/http/statistics
```

### Get Recent Attacks
```bash
curl http://localhost:8080/api/http/attacks?limit=10
```

### Get Uploaded Files
```bash
curl http://localhost:8080/api/http/uploads?limit=20
```

### Get Captured Credentials
```bash
curl http://localhost:8080/api/http/credentials?limit=50
```

---

## Real-time Monitoring

The dashboard uses JavaScript to automatically refresh data:

- **Auto-refresh Interval:** 10 seconds
- **Last Update Time:** Displayed at the top
- **Visual Indicators:** Badges show item counts
- **Color Coding:** Status indicators for quick recognition

---

## Tips

### Monitor Live Activity
1. Open dashboard: http://localhost:8080
2. Switch to "HTTP Honeypot" tab
3. Watch requests appear in real-time

### Track Attacks
1. Go to "Attacks" tab
2. Watch for attack severity colors
3. Review attack payloads

### Identify Scanners
1. Go to "HTTP Honeypot" tab
2. Scroll to "Detected Scanners"
3. See which tools are scanning you

### Review Malware
1. Go to "Uploads" tab
2. Look for ☠️ MALICIOUS files
3. Check MD5 hashes for VirusTotal lookup

---

## Troubleshooting

### Dashboard not loading?
```bash
# Check if web server is running
docker-compose logs honeypot | grep "Web server"

# Restart container
docker-compose restart
```

### No data showing?
```bash
# Generate test data
curl http://localhost/wordpress/wp-login.php
curl -X POST http://localhost/wordpress/wp-login.php -d "log=admin&pwd=admin123"
curl "http://localhost/?id=1' OR 1=1--"

# Wait 10 seconds for auto-refresh
```

### API errors?
```bash
# Test API directly
curl http://localhost:8080/api/http/statistics

# Check logs
docker-compose logs honeypot | grep ERROR
```

---

## Dashboard Features

✅ **Real-time Updates** - Auto-refresh every 10 seconds
✅ **Multi-Protocol** - SSH + HTTP honeypots
✅ **Attack Detection** - 8 attack types detected
✅ **Malware Analysis** - Automatic file detection
✅ **Scanner Fingerprinting** - Identify automated tools
✅ **Credential Capture** - Username/password logging
✅ **Geographic Tracking** - Country-based statistics
✅ **Severity Classification** - Color-coded threats
✅ **Responsive Design** - Works on mobile/desktop
✅ **RESTful APIs** - Integrate with other tools

---

Enjoy monitoring your honeypot! 🍯
