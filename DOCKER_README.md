# Docker Deployment Guide for Honeypot

## 🆕 What's New in v2.0.0

- ✅ **Fixed SSH Command Logging** - Dashboard now shows IP addresses and usernames
- ✅ **Dark Mode Dashboard** - Beautiful theme with ☀️/🌙 toggle
- ✅ **Go Attack Tester** - Test your honeypot from the host (see Testing section)

## Quick Start

### 1. Build and Run with Docker Compose

```bash
# Build and start the honeypot
docker-compose up -d

# View logs
docker-compose logs -f

# Stop the honeypot
docker-compose down

# Stop and remove all data
docker-compose down -v
```

### 2. Manual Docker Build

```bash
# Build the image
docker build -t honeypot:latest .

# Run the container
docker run -d \
  --name honeypot \
  -p 2222:2222 \
  -p 80:80 \
  -p 443:443 \
  -p 8080:8080 \
  -v $(pwd)/logs:/root/logs \
  -v $(pwd)/config.yaml:/root/config.yaml \
  honeypot:latest
```

## Exposed Ports

| Port | Service | Description |
|------|---------|-------------|
| 2222 | SSH | SSH honeypot (fake SSH server) |
| 80 | HTTP | HTTP honeypot (WordPress, phpMyAdmin, etc.) |
| 443 | HTTPS | HTTPS honeypot (if TLS enabled) |
| 8080 | Web UI | Admin dashboard with dark mode 🌙 |

## Accessing the Honeypot

### Admin Dashboard with Dark Mode 🌙
- **URL**: http://localhost:8080
- **Features**:
  - View captured attacks, credentials, uploaded files
  - SSH command history with IP addresses and usernames ✨ (NEW!)
  - HTTP attack detection
  - Dark mode toggle in navbar
  - Real-time statistics

### Test with Go Tester (From Host)

The Go tester runs from your **host machine**, not inside Docker:

```bash
# Build the tester on your host
go build -o tester ./cmd/tester/main.go

# Test the Dockerized honeypot
./tester --ssh-host localhost \
         --ssh-port 2222 \
         --ssh-user root \
         --ssh-pass toor \
         --http-url http://localhost:80

# Or test only SSH
./tester --ssh-only

# Or test only HTTP
./tester --http-only
```

This will execute 50+ attack patterns against your Docker honeypot!

### Manual Testing

#### Test HTTP Honeypot
```bash
# Test WordPress login
curl http://localhost/wordpress/wp-login.php

# Test phpMyAdmin
curl http://localhost/phpmyadmin/

# Test file upload
curl http://localhost/upload.php

# Test SQL injection (will be detected)
curl "http://localhost/wordpress/?id=1' OR '1'='1"
```

#### Test SSH Honeypot
```bash
# Try to connect (use credentials from config.yaml)
ssh root@localhost -p 2222
# Default password: toor

# Try commands once connected
ls
cat /etc/passwd
whoami
```

## Configuration

Edit `config.yaml` before starting:

```yaml
server:
  host: "0.0.0.0"
  port: 2222

web:
  host: "0.0.0.0"
  port: 8080  # Dashboard port

http:
  enabled: true
  port: 80  # HTTP honeypot port

auth:
  fake_users:
    - username: "root"
      password: "toor"
    - username: "admin"
      password: "admin123"
```

## Data Persistence

All data is stored in Docker volumes:
- `honeypot-data`: SQLite database (includes fixed SSH command logging!)
- `honeypot-uploads`: Uploaded files (quarantined malware)
- `./logs`: Application logs

## View Data

```bash
# Access the database to verify SSH commands show IP/username
docker exec -it honey-ssh-honeypot sqlite3 /root/data/honeypot.db \
  "SELECT c.command, cn.remote_addr, cn.username
   FROM commands c
   LEFT JOIN connections cn ON c.connection_id = cn.id
   LIMIT 10;"

# View uploaded files
docker exec -it honey-ssh-honeypot ls -la /root/uploads/

# View logs
docker exec -it honey-ssh-honeypot cat /root/logs/honeypot.log
```

## Dark Mode Dashboard

The dashboard now includes dark mode:

1. Open http://localhost:8080
2. Click the ☀️/🌙 button in the top-right navbar
3. Your theme preference is saved automatically!

**Features:**
- Smooth color transitions
- Persistent across browser sessions
- Optimized readability for both themes
- Works on all dashboard tabs

## Production Deployment

### Security Recommendations

1. **Run on a dedicated server** - Isolate from production systems
2. **Use non-standard ports** - Change default ports in config
3. **Enable TLS** - Configure certificates for HTTPS
4. **Monitor actively** - Set up email alerts
5. **Regular backups** - Backup the database regularly
6. **Restrict dashboard access** - Bind to localhost only

### Example Production docker-compose.yml

```yaml
version: '3.8'

services:
  honeypot:
    build: .
    container_name: honeypot-prod
    ports:
      - "22:2222"     # Make it look like real SSH
      - "80:80"
      - "443:443"
      - "127.0.0.1:8080:8080"  # Only localhost can access dashboard
    volumes:
      - ./logs:/root/logs
      - honeypot-data:/root/data
      - honeypot-uploads:/root/uploads
      - ./config.yaml:/root/config.yaml
      - ./certs:/etc/honeypot  # TLS certificates
    restart: always
    networks:
      - honeypot-network
    cap_add:
      - NET_BIND_SERVICE

networks:
  honeypot-network:
    driver: bridge

volumes:
  honeypot-data:
  honeypot-uploads:
```

## Testing Your Deployment

### 1. Start the Honeypot
```bash
docker-compose up -d
```

### 2. Verify Services
```bash
# Check if all ports are listening
docker exec honey-ssh-honeypot netstat -tlnp

# Should show:
# 0.0.0.0:2222  (SSH honeypot)
# 0.0.0.0:80    (HTTP honeypot)
# 0.0.0.0:8080  (Dashboard)
```

### 3. Run Attack Tests
```bash
# From your host machine
./tester

# You should see:
# ✓ SSH brute force attempts
# ✓ SSH command executions
# ✓ HTTP attacks (SQL injection, XSS, etc.)
# ✓ File uploads
# ✓ Scanner detection
```

### 4. Verify Dashboard
```bash
# Open browser
http://localhost:8080

# Check:
# ✓ SSH tab shows commands with IP addresses
# ✓ HTTP tab shows attack detections
# ✓ Dark mode toggle works
# ✓ Statistics are updating
```

## Monitoring

### Real-time Logs
```bash
docker-compose logs -f honeypot
```

### Check Running Services
```bash
docker-compose ps
```

### Resource Usage
```bash
docker stats honey-ssh-honeypot
```

## Troubleshooting

### Port Already in Use
```bash
# Check what's using port 80
sudo lsof -i :80

# Change port in docker-compose.yml
ports:
  - "8000:80"  # Use port 8000 instead
```

### Dashboard Not Accessible
```bash
# Check if port 8080 is exposed
docker ps | grep honeypot

# Should show: 0.0.0.0:8080->8080/tcp

# If not, update docker-compose.yml:
ports:
  - "8080:8080"
```

### SSH Commands Still Show "-" for IP/Username
```bash
# This means you're using old database data
# Delete the database and restart:

docker-compose down
docker volume rm honeypot_honeypot-data
docker-compose up -d

# Then run tests again:
./tester
```

### Dark Mode Not Saving
```bash
# This is a browser localStorage issue
# Clear browser cache or try:
# - Different browser
# - Incognito/Private window
# - Check browser console for errors
```

### Permission Denied on Port 80/443
```bash
# Ensure cap_add is set in docker-compose.yml
cap_add:
  - NET_BIND_SERVICE
```

### Database Locked
```bash
# Stop the container
docker-compose down

# Remove the database volume
docker volume rm honeypot_honeypot-data

# Restart
docker-compose up -d
```

## Updating to v2.0.0

If upgrading from v1.0.0:

```bash
# Pull latest changes
git pull

# IMPORTANT: Delete old database to fix SSH logging
docker-compose down
docker volume rm honeypot_honeypot-data

# Rebuild with new features
docker-compose build --no-cache

# Start with fresh database
docker-compose up -d

# Test the new features
./tester
```

## Backup and Restore

### Backup
```bash
# Backup database (includes all attack data)
docker cp honey-ssh-honeypot:/root/data/honeypot.db ./backup_$(date +%Y%m%d).db

# Backup uploaded files
docker cp honey-ssh-honeypot:/root/uploads ./uploads_backup_$(date +%Y%m%d)

# Backup logs
docker cp honey-ssh-honeypot:/root/logs ./logs_backup_$(date +%Y%m%d)
```

### Restore
```bash
# Restore database
docker cp ./backup_20241220.db honey-ssh-honeypot:/root/data/honeypot.db

# Restart container
docker-compose restart
```

## Advanced: Multi-Instance Deployment

Deploy multiple honeypots on different ports:

```yaml
# docker-compose.multi.yml
version: '3.8'

services:
  honeypot-ssh:
    build: .
    container_name: honeypot-ssh
    ports:
      - "2222:2222"
      - "8080:8080"
    volumes:
      - ./config-ssh.yaml:/root/config.yaml
      - honeypot-ssh-data:/root/data
    restart: unless-stopped

  honeypot-http:
    build: .
    container_name: honeypot-http
    ports:
      - "80:80"
      - "8081:8080"
    volumes:
      - ./config-http.yaml:/root/config.yaml
      - honeypot-http-data:/root/data
    restart: unless-stopped

volumes:
  honeypot-ssh-data:
  honeypot-http-data:
```

Start with:
```bash
docker-compose -f docker-compose.multi.yml up -d
```

## Performance Tuning

### Resource Limits
```yaml
# Add to docker-compose.yml
services:
  honeypot:
    # ... existing config ...
    deploy:
      resources:
        limits:
          cpus: '0.5'
          memory: 256M
        reservations:
          cpus: '0.25'
          memory: 128M
```

### Log Rotation
```yaml
services:
  honeypot:
    # ... existing config ...
    logging:
      driver: "json-file"
      options:
        max-size: "10m"
        max-file: "3"
```

## Network Security

### Firewall Rules (Production)
```bash
# Only allow specific IPs to access dashboard
sudo ufw allow from 192.168.1.0/24 to any port 8080

# Block dashboard from internet
sudo ufw deny 8080/tcp
```

### Reverse Proxy (Nginx)
```nginx
# /etc/nginx/sites-available/honeypot-dashboard
server {
    listen 443 ssl;
    server_name honeypot.example.com;

    ssl_certificate /etc/letsencrypt/live/honeypot.example.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/honeypot.example.com/privkey.pem;

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

## Health Checks

Add to docker-compose.yml:
```yaml
services:
  honeypot:
    # ... existing config ...
    healthcheck:
      test: ["CMD", "netstat", "-tlnp", "|", "grep", "2222"]
      interval: 30s
      timeout: 10s
      retries: 3
```

## Support

### Check Logs
```bash
docker-compose logs --tail=100 honeypot
```

### Debug Mode
```bash
# Run with debug output
docker-compose up

# Or attach to running container
docker attach honey-ssh-honeypot
```

### Common Issues

| Issue | Solution |
|-------|----------|
| Dashboard shows old data | Delete database volume and restart |
| SSH commands show "-" | Using old database, see "Updating to v2.0.0" |
| Dark mode doesn't work | Clear browser cache, try incognito |
| Tester can't connect | Check firewall, verify ports are exposed |
| Container keeps restarting | Check logs: `docker logs honey-ssh-honeypot` |

## Additional Resources

- **README_GO_TESTER.md** - Complete tester documentation
- **DARK_MODE_GUIDE.md** - Dark mode customization
- **QUICK_START.md** - Quick start guide for non-Docker deployment
- **CHANGELOG.md** - Version history

---

**Everything working?** Your Dockerized honeypot is ready with dark mode and proper attack tracking! 🐳🍯✨
