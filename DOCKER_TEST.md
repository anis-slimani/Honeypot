# Docker Quick Test

## Build and Start
```bash
docker-compose up -d
```

## Verify Ports
```bash
docker ps
# Should show:
# 0.0.0.0:2222->2222/tcp  (SSH honeypot)
# 0.0.0.0:80->80/tcp      (HTTP honeypot)
# 0.0.0.0:8080->8080/tcp  (Dashboard)
```

## Access Dashboard
```
http://localhost:8080
```

## Test from Host
```bash
# Build tester
go build -o tester ./cmd/tester/main.go

# Test containerized honeypot
./tester --ssh-host localhost \
         --ssh-port 2222 \
         --http-url http://localhost:80
```

## Verify Features
- ✅ SSH commands show IP & username
- ✅ Dark mode toggle works
- ✅ HTTP attacks detected
