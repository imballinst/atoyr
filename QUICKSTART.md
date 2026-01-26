# Quick Start Guide - Atoyr (Go Backend)

## Prerequisites

- Go 1.21 or later
- Node.js 18+ (for client)
- Yarn package manager

## Getting Started

### 1. Install Dependencies

```bash
cd /workspaces/atoyr

# Install Go dependencies (auto-fetched with go commands)
# Install Node dependencies for client
yarn install
```

### 2. Build Go Server

```bash
cd packages/server

# Option 1: Using Make
make build

# Option 2: Using Go directly
go build -o bin/server ./cmd
```

### 3. Run Server

```bash
cd packages/server

# Option 1: Development mode (watch + rebuild)
make run

# Option 2: Run compiled binary
./bin/server

# Option 3: Direct Go execution
go run ./cmd
```

Server starts on `http://localhost:3000`

### 4. Run Tests

```bash
cd packages/server

# Option 1: All tests
make test

# Option 2: Go directly
go test -v ./...

# Option 3: With coverage
go test -v -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### 5. Start Client (Separate Terminal)

```bash
cd packages/client
yarn dev
```

Client runs on `http://localhost:5173`

## Full Stack Development

### Run Both Server and Client

```bash
# From root workspace
yarn dev

# This runs:
# - Client: yarn --cwd packages/client dev (port 5173)
# - Server: go run ./packages/server/cmd (port 3000)
```

### Build Production

```bash
# From root workspace
yarn build

# This:
# - Builds client to packages/client/dist
# - Builds server to packages/server/bin/server
```

## Project Structure

```
packages/server/
├── cmd/main.go                    # Entry point
├── internal/
│   ├── database/database.go       # DB initialization
│   ├── models/models.go           # Data models
│   ├── services/                  # Business logic
│   │   ├── game.service.go
│   │   ├── session.service.go
│   │   ├── word.service.go
│   │   └── leaderboard.service.go
│   ├── routes/                    # HTTP handlers
│   │   ├── game_routes.go
│   │   └── leaderboard_routes.go
│   └── middleware/                # CORS, etc.
├── go.mod                         # Dependencies
├── Makefile                       # Build targets
└── README.md                      # Full documentation
```

## Common Commands

```bash
# Clean builds
cd packages/server && make clean && make build

# Run with debug logs
NODE_ENV=development go run ./cmd

# Test with specific service
cd packages/server && go test -v ./internal/services/game_test.go

# Format code
cd packages/server && go fmt ./...

# Lint (requires golangci-lint)
cd packages/server && golangci-lint run ./...
```

## Environment Variables

```bash
# Server (packages/server)
export NODE_ENV=development      # or production
export PORT=3000                 # HTTP port
export DATABASE_PATH=/tmp/game.db # SQLite path
```

## API Testing

### Start Game

```bash
curl -X POST http://localhost:3000/api/game/start \
  -H "Content-Type: application/json" \
  -d '{"autoVoice": true}'
```

Response:

```json
{
  "sessionId": "uuid-here",
  "currentWord": "scrambled_word",
  "token": "sha256_hash"
}
```

### Submit Answer

```bash
curl -X POST http://localhost:3000/api/game/answer \
  -H "Content-Type: application/json" \
  -d '{
    "sessionId": "uuid-here",
    "answer": "unscrambled_word"
  }'
```

### Get Leaderboard

```bash
curl http://localhost:3000/api/leaderboard?page=0&limit=10
```

### Server-Sent Events

```bash
curl http://localhost:3000/api/game/sse/uuid-here
```

## Troubleshooting

### Go module not found

```bash
cd packages/server
go mod tidy
go mod download
```

### Port already in use

```bash
# Change PORT
export PORT=3001
go run ./cmd

# Or kill existing process
lsof -i :3000
kill -9 <PID>
```

### Database file permission denied

```bash
# Ensure directory exists and is writable
mkdir -p ~/.atoyr
chmod 755 ~/.atoyr
```

### Words file not found

```bash
# Verify file exists
ls -la packages/client/src/data/words.json

# Or specify custom path
export WORDS_PATH=/path/to/words.json
```

## Next Steps

1. **Explore API** - Try the endpoints with curl or Postman
2. **Run Tests** - `go test -v ./...` to verify everything
3. **Read Code** - Check `internal/services/game.service.go` for core logic
4. **Deploy** - See `packages/server/README.md` for deployment options
5. **Customize** - Add auth, caching, metrics as needed

## Documentation

- **Server**: See `packages/server/README.md`
- **Architecture**: See `.opencode/specs/02-backend.md`
- **Migration**: See `.opencode/specs/03-go-migration.md`
- **Client**: See `packages/client/README.md`

## Getting Help

```bash
# Check Go version
go version

# See available Go commands
go help

# Check server logs
NODE_ENV=development go run ./cmd 2>&1

# Run specific test with verbose output
go test -v ./internal/services/game_test.go -run TestGameService_StartGame
```

Enjoy building Atoyr! 🎮
