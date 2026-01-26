# Deployment Guide - Atoyr Backend (Go)

## Pre-Deployment Checklist

- [ ] All tests passing: `go test -v ./...`
- [ ] Binary builds successfully: `go build -o bin/server ./cmd`
- [ ] Environment variables configured
- [ ] Database path is writable
- [ ] Firewall ports open (3000 or custom PORT)
- [ ] Client built and configured to point to backend URL

## Deployment Options

### 1. Docker (Recommended for Production)

#### Building Docker Image

**Multi-stage Dockerfile (Production-optimized):**

```dockerfile
# Stage 1: Build
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install dependencies
RUN apk add --no-cache git sqlite-dev

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY cmd/ cmd/
COPY internal/ internal/

# Build binary
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o bin/server ./cmd

# Stage 2: Runtime
FROM alpine:latest

RUN apk --no-cache add ca-certificates sqlite-libs

WORKDIR /root/

# Copy binary from builder
COPY --from=builder /app/bin/server .

# Create data directory
RUN mkdir -p /root/.atoyr

EXPOSE 3000

CMD ["./server"]
```

**Build and Run:**

```bash
# Build image
docker build -t atoyr-server:latest .

# Run container
docker run -d \
  --name atoyr-server \
  -p 3000:3000 \
  -e NODE_ENV=production \
  -e PORT=3000 \
  -v ~/.atoyr:/root/.atoyr \
  atoyr-server:latest

# Check logs
docker logs -f atoyr-server

# Stop container
docker stop atoyr-server
docker rm atoyr-server
```

### 2. Systemd Service (Linux)

**Create `/etc/systemd/system/atoyr-server.service`:**

```ini
[Unit]
Description=Atoyr Server (Go)
After=network.target

[Service]
Type=simple
User=atoyr
WorkingDirectory=/opt/atoyr
ExecStart=/opt/atoyr/bin/server
Restart=on-failure
RestartSec=5

# Environment
Environment="NODE_ENV=production"
Environment="PORT=3000"
Environment="DATABASE_PATH=/opt/atoyr/data/atoyr.sqlite"

# Resource limits
LimitNOFILE=65535
MemoryLimit=512M

# Security
NoNewPrivileges=true
PrivateTmp=true

[Install]
WantedBy=multi-user.target
```

**Deploy:**

```bash
# Create user and directories
sudo useradd -r -s /bin/false atoyr
sudo mkdir -p /opt/atoyr/data
sudo chown -R atoyr:atoyr /opt/atoyr

# Copy binary
sudo cp bin/server /opt/atoyr/
sudo chown atoyr:atoyr /opt/atoyr/bin/server
sudo chmod +x /opt/atoyr/bin/server

# Enable and start service
sudo systemctl daemon-reload
sudo systemctl enable atoyr-server
sudo systemctl start atoyr-server

# Monitor
sudo journalctl -u atoyr-server -f
```

### 3. Cloud Platforms

#### AWS EC2

```bash
# SSH into instance
ssh -i key.pem ec2-user@instance-ip

# Install Go (if needed)
wget https://golang.org/dl/go1.21.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin

# Deploy binary
scp -i key.pem bin/server ec2-user@instance-ip:~/
ssh -i key.pem ec2-user@instance-ip 'chmod +x ~/server && ~/server'

# Or use Docker/ECS
```

#### Heroku

```bash
# Create app
heroku create atoyr-server

# Set buildpack
heroku buildpacks:set heroku/go

# Deploy
git push heroku main

# Check logs
heroku logs --tail
```

Create `Procfile`:

```
web: bin/server
```

#### Google Cloud Run

```bash
# Build and push image
gcloud builds submit --tag gcr.io/PROJECT_ID/atoyr-server

# Deploy
gcloud run deploy atoyr-server \
  --image gcr.io/PROJECT_ID/atoyr-server \
  --platform managed \
  --memory 256Mi \
  --timeout 3600s \
  --set-env-vars NODE_ENV=production
```

#### DigitalOcean App Platform

```bash
# Deploy from GitHub
# 1. Connect GitHub repo in DigitalOcean console
# 2. Select branch and directory (packages/server)
# 3. Configure environment variables
# 4. Deploy
```

### 4. Binary Deployment (Simple)

**SSH Deploy:**

```bash
# Compile on local machine
cd packages/server
make build

# SCP to server
scp bin/server user@server:/opt/atoyr/

# SSH and run
ssh user@server
cd /opt/atoyr
chmod +x server
./server
```

**For permanent background execution:**

```bash
# Use screen
screen -S atoyr-server
./server
# Press Ctrl-A then D to detach

# Use nohup
nohup ./server > atoyr.log 2>&1 &
tail -f atoyr.log
```

## Configuration

### Environment Variables

```bash
# Server
export NODE_ENV=production              # development or production
export PORT=3000                        # HTTP port
export DATABASE_PATH=/opt/atoyr/data    # SQLite database location

# Optional
export WORDS_PATH=/opt/atoyr/words.json # Custom words file
```

### Database Setup

```bash
# Create data directory
mkdir -p ~/.atoyr
chmod 755 ~/.atoyr

# Server will auto-create SQLite database on first run
# Database file: ~/.atoyr/atoyr.sqlite
```

### SSL/TLS (Reverse Proxy)

**Nginx configuration:**

```nginx
upstream atoyr_backend {
  server localhost:3000;
}

server {
  listen 80;
  server_name api.atoyr.com;
  return 301 https://$server_name$request_uri;
}

server {
  listen 443 ssl http2;
  server_name api.atoyr.com;

  ssl_certificate /etc/letsencrypt/live/api.atoyr.com/fullchain.pem;
  ssl_certificate_key /etc/letsencrypt/live/api.atoyr.com/privkey.pem;

  location / {
    proxy_pass http://atoyr_backend;
    proxy_http_version 1.1;
    proxy_set_header Upgrade $http_upgrade;
    proxy_set_header Connection "upgrade";
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;

    # For SSE
    proxy_buffering off;
    proxy_cache off;
  }
}
```

## Monitoring & Maintenance

### Health Check

```bash
# Simple health check endpoint
curl http://localhost:3000/api/leaderboard

# Expected: 200 OK with JSON response
```

**Add health endpoint (optional):**

```go
// In main.go
router.GET("/health", func(c *gin.Context) {
  c.JSON(200, gin.H{"status": "ok"})
})
```

### Logging

**Development (verbose):**

```bash
NODE_ENV=development go run ./cmd
```

**Production (quiet):**

```bash
NODE_ENV=production go run ./cmd
```

### Database Backup

```bash
# Backup SQLite database
cp ~/.atoyr/atoyr.sqlite ~/.atoyr/atoyr.sqlite.backup

# Or automated backup
# Add cron job:
# 0 2 * * * cp ~/.atoyr/atoyr.sqlite ~/backups/atoyr.sqlite.$(date +\%Y\%m\%d)
```

### Performance Monitoring

```bash
# Monitor memory and CPU
top -p $(pgrep -f "bin/server")

# Or with Go's built-in metrics (if added)
curl http://localhost:3000/metrics
```

### Crash Recovery

**Systemd handles automatic restart:**

```bash
# Check if running
sudo systemctl status atoyr-server

# Manual restart
sudo systemctl restart atoyr-server

# Check recent errors
sudo journalctl -u atoyr-server -n 50
```

## Scaling Considerations

### Horizontal Scaling

Each binary is stateless:

```
┌──────────────┐
│   Client     │
└──────┬───────┘
       │
   ┌───┴─────┬──────────┬──────────┐
   │          │          │          │
┌──▼──┐  ┌──▼──┐  ┌──▼──┐  ┌──▼──┐
│Inst1│  │Inst2│  │Inst3│  │Inst4│ (Load Balanced)
└──┬──┘  └──┬──┘  └──┬──┘  └──┬──┘
   │        │        │        │
   └────────┴────────┴────────┘
           │
        ┌──▼────┐
        │ SQLite│ (Shared)
        └───────┘
```

**Load Balancer Setup (Nginx):**

```nginx
upstream atoyr_servers {
  least_conn;
  server backend1:3000;
  server backend2:3000;
  server backend3:3000;
  server backend4:3000;
}

server {
  listen 80;
  location / {
    proxy_pass http://atoyr_servers;
  }
}
```

### Database Considerations

- SQLite works for up to ~100k records
- For larger scale: migrate to PostgreSQL
- Schema is identical, only connection string changes

### Caching Layer (Optional)

Add Redis for leaderboard caching:

```bash
docker run -d -p 6379:6379 redis:latest
```

Cache leaderboard queries:

- TTL: 60 seconds (refresh every minute)
- Cache key: `leaderboard:page:{page}:limit:{limit}`

## Troubleshooting

### Port Already in Use

```bash
# Find process using port 3000
lsof -i :3000

# Kill process
kill -9 <PID>

# Or use different port
export PORT=3001
./server
```

### Database Lock

```bash
# Remove lock file if corrupted
rm ~/.atoyr/atoyr.sqlite-wal
rm ~/.atoyr/atoyr.sqlite-shm

# Restart server
./server
```

### Connection Refused

```bash
# Check if server is running
curl http://localhost:3000/api/leaderboard

# Check firewall
sudo ufw status
sudo ufw allow 3000/tcp

# Check binding
netstat -tulpn | grep 3000
```

### Out of Memory

```bash
# Increase system memory
# Or limit process:
ulimit -m 512000  # 512MB max

# In systemd service:
MemoryLimit=512M
```

## Rollback Procedure

If issues occur:

```bash
# Kill current server
pkill -f "bin/server"

# Restore from backup
cp ~/.atoyr/atoyr.sqlite.backup ~/.atoyr/atoyr.sqlite

# Start old binary
./bin/server.backup
```

## Post-Deployment Validation

```bash
# 1. Test all endpoints
./test_endpoints.sh

# 2. Check database is populated
sqlite3 ~/.atoyr/atoyr.sqlite "SELECT COUNT(*) FROM result_entities;"

# 3. Monitor logs for errors
tail -f /var/log/atoyr-server.log

# 4. Test with client
open http://localhost:5173

# 5. Play a game to verify end-to-end
```

## Continuous Deployment (GitHub Actions)

Add to workflow to auto-deploy on push to main:

```yaml
deploy:
  runs-on: ubuntu-latest
  needs: test
  if: github.ref == 'refs/heads/main'

  steps:
    - uses: actions/checkout@v4
    - uses: actions/setup-go@v4

    - name: Build
      run: cd packages/server && go build -o bin/server ./cmd

    - name: Deploy to server
      env:
        DEPLOY_KEY: ${{ secrets.DEPLOY_KEY }}
      run: |
        mkdir -p ~/.ssh
        echo "$DEPLOY_KEY" > ~/.ssh/deploy_key
        chmod 600 ~/.ssh/deploy_key
        scp -i ~/.ssh/deploy_key bin/server deploy@server:/opt/atoyr/
        ssh -i ~/.ssh/deploy_key deploy@server 'sudo systemctl restart atoyr-server'
```

---

## Support & Resources

- **Documentation**: `packages/server/README.md`
- **Quick Start**: `QUICKSTART.md`
- **Issues**: Check GitHub issues
- **Logs**: `journalctl` (systemd) or `docker logs` (Docker)

---

**Happy Deploying! 🚀**
