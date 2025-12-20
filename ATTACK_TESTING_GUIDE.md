# Honeypot Attack Testing Guide

## What Was Fixed

### SSH Command Logging Issue
**Problem:** The dashboard was showing IP addresses and usernames as "-" for executed SSH commands.

**Root Cause:** The `SaveConnection` function in `internal/database/database.go` was not returning the inserted connection ID. This meant that when commands were saved, they were being saved with `connection_id = 0`, which didn't match any real connection in the database. When the web dashboard performed a LEFT JOIN to get the IP address and username, it returned NULL values.

**Solution:**
1. Modified `SaveConnection()` to return the inserted connection ID
2. Updated SSH server code to use the returned connection ID
3. Updated fake session handler to properly store the connection ID

### Files Modified
- `internal/database/database.go` - SaveConnection now returns (int, error)
- `internal/honeypot/ssh_server.go` - Updated to handle returned ID
- `internal/honeypot/fake_session.go` - Updated to store connection ID correctly

## Attack Testing Script

### Overview
The `test_attacks.py` script is a comprehensive testing tool that simulates various attacks against both SSH and HTTP honeypots.

### Prerequisites
```bash
pip3 install paramiko requests
```

### Usage

#### Test All Attacks (Default)
```bash
python3 test_attacks.py
```

#### Test Only SSH Attacks
```bash
python3 test_attacks.py --ssh-only \
    --ssh-host localhost \
    --ssh-port 2222 \
    --ssh-user root \
    --ssh-pass toor
```

#### Test Only HTTP Attacks
```bash
python3 test_attacks.py --http-only \
    --http-url http://localhost:8080
```

#### Custom Configuration
```bash
python3 test_attacks.py \
    --ssh-host 192.168.1.100 \
    --ssh-port 2222 \
    --ssh-user admin \
    --ssh-pass admin123 \
    --http-url http://192.168.1.100:8080
```

#### Skip Brute Force Tests
```bash
python3 test_attacks.py --no-bruteforce
```

### Attack Types Tested

#### SSH Attacks
1. **Brute Force Attack** - Tests multiple password combinations
2. **Valid Login** - Tests with configured fake credentials
3. **Command Execution** - Tests various Unix commands including:
   - Basic commands (ls, pwd, whoami)
   - Reconnaissance (uname, ps, netstat)
   - File access (/etc/passwd, /etc/shadow)
   - Malicious downloads (wget, curl)
   - Destructive commands (rm -rf)

#### HTTP Attacks

##### SQL Injection
- UNION-based injection
- Boolean-based blind injection
- Comment-based injection
- Information schema queries

##### Cross-Site Scripting (XSS)
- Reflected XSS
- DOM-based XSS
- Event handler injection
- SVG-based XSS

##### Command Injection
- Shell command chaining
- Pipe-based injection
- Backtick execution
- Remote shell attempts

##### Path Traversal
- Directory traversal (../)
- Encoded traversal (%2f)
- Windows path traversal
- Config file access attempts

##### Malicious File Upload
- PHP web shells
- JSP backdoors
- Fake malware executables
- Multiple file extensions

##### Application-Specific Attacks
- WordPress login brute force
- phpMyAdmin credential attacks
- Admin panel discovery

##### Scanner Detection
- Simulates common security scanners:
  - Nikto
  - SQLMap
  - Nmap NSE
  - ZmEu
  - Masscan

##### Common Exploits
- CGI script exploitation
- IIS Unicode attacks
- Environment file disclosure
- Git repository access
- Backup file access

## Testing Workflow

### 1. Start the Honeypot
```bash
# Build and run
go build -o honeypot .
sudo ./honeypot

# Or with Docker
docker-compose up -d
```

### 2. Run Attack Tests
```bash
# Make sure you have valid credentials configured in config.yaml
# Then run the test script
python3 test_attacks.py
```

### 3. View Results
Open the dashboard in your browser:
```
http://localhost:9090
```

Navigate to:
- **SSH Tab** → Command Execution History
- **HTTP Tab** → Attack Detection
- **Attacks Tab** → Threat Intelligence

### 4. Verify IP and Username Capture
In the Command Execution History table, you should now see:
- ✅ IP Address column populated with attacker IPs
- ✅ Username column showing the authenticated user
- ✅ Command column with executed commands
- ✅ Timestamp column with execution time

## Expected Behavior

### Before the Fix
```
IP ADDRESS  USERNAME  COMMAND           TIMESTAMP
-           -         ls                12/16/2025, 8:49:34 PM
-           -         cat hostname      12/16/2025, 8:49:20 PM
```

### After the Fix
```
IP ADDRESS       USERNAME  COMMAND           TIMESTAMP
192.168.1.50     root      ls                12/16/2025, 8:49:34 PM
192.168.1.50     root      cat hostname      12/16/2025, 8:49:20 PM
10.0.0.15        admin     whoami            12/16/2025, 8:50:12 PM
```

## Configuration

### Fake SSH Users
Edit `config.yaml` to add fake users that will accept logins:

```yaml
auth:
  fake_users:
    - username: "root"
      password: "toor"
    - username: "admin"
      password: "admin123"
    - username: "user"
      password: "password"
```

### HTTP Honeypot Settings
```yaml
http:
  enabled: true
  port: 8080
  wordpress_enabled: true
  phpmyadmin_enabled: true
  upload_enabled: true
```

## Troubleshooting

### SSH Tests Failing
1. Check if SSH honeypot is running: `netstat -tlnp | grep 2222`
2. Verify credentials in config.yaml match test script arguments
3. Check logs: `tail -f /var/log/honeypot.log`

### HTTP Tests Failing
1. Check if HTTP server is running: `curl http://localhost:8080`
2. Verify port in test script matches config
3. Check for firewall rules blocking connections

### Commands Not Showing IP/Username
1. Stop the honeypot
2. Delete the database: `rm honeypot.db`
3. Rebuild: `go build -o honeypot .`
4. Start fresh and test again

## Advanced Testing

### Load Testing
```bash
# Run multiple instances in parallel
for i in {1..5}; do
    python3 test_attacks.py &
done
wait
```

### Continuous Testing
```bash
# Run tests every 5 minutes
watch -n 300 python3 test_attacks.py
```

### Custom Attack Sequences
Edit `test_attacks.py` to add your own attack patterns:

```python
def test_custom_attack(base_url):
    """Test custom attack pattern"""
    print_section("CUSTOM ATTACK")

    # Your custom test here
    r = requests.get(f"{base_url}/your-path")
    print_test("Custom test", True, f"Status: {r.status_code}")
```

## Security Notes

⚠️ **WARNING:** This script generates real attack traffic. Only use it against:
- Your own honeypot systems
- Systems you have permission to test
- Isolated lab environments

**DO NOT** run against:
- Production systems
- Systems you don't own
- Public infrastructure
- Without explicit authorization

## Monitoring Tips

### Real-time Log Monitoring
```bash
# Watch all honeypot activity
tail -f /var/log/honeypot.log

# Filter for SSH commands
tail -f /var/log/honeypot.log | grep "Command from"

# Filter for HTTP attacks
tail -f /var/log/honeypot.log | grep "Attack detected"
```

### Database Queries
```bash
sqlite3 honeypot.db

# View recent commands with IP and user
SELECT c.command, c.executed_at, cn.remote_addr, cn.username
FROM commands c
LEFT JOIN connections cn ON c.connection_id = cn.id
ORDER BY c.executed_at DESC
LIMIT 10;

# View attack statistics
SELECT attack_type, COUNT(*) as count
FROM http_attacks
GROUP BY attack_type
ORDER BY count DESC;
```

## Contributing

To add new attack tests:
1. Create a new test function in `test_attacks.py`
2. Add it to the main() function
3. Document the attack type in this guide
4. Submit a pull request

## License

This testing framework is provided as-is for security research and testing purposes.
