# Atoyr Server (Go + Gin + GORM)

A high-performance backend for the Atoyr word game, built with Go using Gin framework and GORM ORM.

## Architecture

### Technology Stack

- **Language**: Go 1.21+
- **Web Framework**: Gin 1.9.1
- **ORM**: GORM 1.25.5
- **Database**: SQLite
- **UUID Generation**: google/uuid v1.6.0

### Project Structure

```
packages/server/
├── cmd/
│   └── main.go                 # Application entry point
├── internal/
│   ├── database/
│   │   └── database.go        # Database initialization
│   ├── middleware/
│   │   └── cors.go            # CORS middleware
│   ├── models/
│   │   └── models.go          # Database entities
│   ├── routes/
│   │   ├── game_routes.go     # Game endpoints
│   │   └── leaderboard_routes.go  # Leaderboard endpoints
│   └── services/
│       ├── game.service.go    # Game logic
│       ├── leaderboard.service.go  # Leaderboard queries
│       ├── session.service.go # Session management
│       ├── word.service.go    # Word selection
│       └── *_test.go          # Unit tests
├── go.mod                      # Go module definition
├── go.sum                      # Dependency checksums
└── Makefile                    # Build targets

```

## API Endpoints

### Game

#### POST /api/game/start

Start a new game session.

**Request:**

```json
{
  "autoVoice": boolean
}
```

**Response:**

```json
{
  "sessionId": "uuid",
  "currentWord": "scrambled-word",
  "token": "sha256-token"
}
```

#### POST /api/game/answer

Submit an answer to the current word.

**Request:**

```json
{
  "sessionId": "uuid",
  "answer": "unscrambled-word"
}
```

**Response:**

```json
{
  "correct": boolean,
  "score": number,
  "attempts": number,
  "remaining": number,
  "newWord": "next-word (if correct)"
}
```

#### GET /api/game/sse/:sessionId

Server-Sent Events for real-time game updates.

**Events:**

- `start` - Game initialization with first word
- `tick` - Per-second updates (remaining time, score, phase)
- `finish` - Game completion with final stats

### Leaderboard

#### GET /api/leaderboard?page=0&limit=10

Get leaderboard entries with pagination.

**Response:**

```json
{
  "entries": [
    {
      "rank": 1,
      "score": 500,
      "totalAttempts": 20,
      "accuracy": 95.5,
      "timestamp": 1234567890
    }
  ],
  "total": 250
}
```

## Database Schema

### SessionEntity

- **ID** (PK): UUID string
- **CreatedAt**: Timestamp
- **ExpiresAt**: Timestamp (5 minutes from creation)
- **Phase**: "waiting-for-opponent" | "playing" | "finished"
- **Score**: Accumulated points (sum of remaining seconds per correct answer)
- **TotalAttempts**: Number of attempts made
- **RemainingSeconds**: Countdown timer (30 seconds per word)
- **AutoVoice**: Boolean flag for audio option
- **UsedWords**: JSON array of completed words
- **CurrentWord**: Current word being guessed
- **CurrentWordToken**: SHA256 hash of current word

### ResultEntity

- **ID** (PK): UUID string
- **SessionID** (FK): References SessionEntity
- **Timestamp**: When game finished
- **Score**: Final score
- **TotalAttempts**: Number of guesses made
- **Accuracy**: Percentage of correct answers
- **DurationMs**: Total game duration in milliseconds

## Services

### SessionService

CRUD operations and state management for game sessions.

**Methods:**

- `Create(autoVoice bool)` - Initialize new session
- `FindByID(id string)` - Retrieve session
- `Update(session *SessionEntity)` - Persist changes
- `SetCurrentWord(sessionID, word, token string)` - Set active word
- `AddUsedWord(sessionID, word string)` - Track completed words
- `UpdatePhase(sessionID, phase string)` - Change game state
- `UpdateScore(sessionID string, scoreIncrement int)` - Add points
- `IncrementTotalAttempts(sessionID string)` - Count attempts
- `SaveResult(result *ResultEntity)` - Store final result
- `UpdateRemainingSeconds(sessionID string, seconds int32)` - Update timer

### WordService

Word selection and management.

**Methods:**

- `GetRandomWord(excludeWords []string)` - Select next word
- Case-insensitive exclusion matching

### GameService

Core game logic and answer validation.

**Methods:**

- `StartGame(sessionID string)` - Initialize gameplay
- `EmitWord(sessionID string)` - Provide next word
- `SubmitAnswer(sessionID, answer string)` - Validate guess
- `FinishGame(sessionID string)` - End game and save result
- `generateToken(word string)` - SHA256 hash for validation
- Background timer goroutine for countdown

### LeaderboardService

Ranking and statistics queries.

**Methods:**

- `GetLeaderboard(limit, offset int)` - Paginated leaderboard
- `GetTopScores(limit int)` - Top N scores
- `GetTotalEntries()` - Total result count

## Building & Running

### Prerequisites

- Go 1.21 or later
- Gin, GORM, and other dependencies (auto-installed with `go mod tidy`)

### Build

```bash
make build
# or
go build -o bin/server ./cmd
```

### Run Development

```bash
make run
# or
go run ./cmd
```

### Run Tests

```bash
make test
# or
go test -v ./...
```

### Clean

```bash
make clean
```

## Environment Variables

- `NODE_ENV` - "development" or "production" (default: "development")
- `PORT` - HTTP server port (default: 3000)
- `DATABASE_PATH` - SQLite database file path (default: ~/.atoyr/atoyr.sqlite)
- `WORDS_PATH` - Path to words.json file (default: ../client/src/data/words.json)

## Testing

Comprehensive test suite with:

- **Unit Tests** - Service logic validation
- **Integration Tests** - API endpoint testing
- **Test Database** - In-memory SQLite for isolation

### Running Tests

```bash
go test -v ./...                    # All tests
go test -v ./internal/services/...  # Service tests only
go test -v ./internal/routes/...    # Route tests only
```

### Test Coverage

```bash
go test -v -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Deployment

### Docker (Optional)

```dockerfile
FROM golang:1.21 AS builder
WORKDIR /app
COPY . .
RUN go build -o bin/server ./cmd

FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=builder /app/bin/server /server
EXPOSE 3000
CMD ["/server"]
```

### Environment Setup

Create `.env` file or export variables:

```bash
export NODE_ENV=production
export PORT=3000
export DATABASE_PATH=/var/lib/atoyr/atoyr.sqlite
```

## Performance

- **Lightweight**: Single binary, minimal dependencies
- **Concurrent**: Goroutine-based timer per session
- **Efficient**: SQLite for simplicity, GORM for ORM operations
- **Scalable**: Stateless design, no session affinity required

## Game Logic

1. **Session Creation** - UUID-based session with 5-minute expiration
2. **Word Emission** - Random word selection with scrambling
3. **Token Validation** - SHA256 hash verification for answer checking
4. **Scoring** - Points based on remaining seconds (0-30)
5. **Accuracy Calculation** - Percentage of correct attempts
6. **Result Persistence** - Final stats saved to leaderboard

## Error Handling

- Standard Go error returns
- Gin error JSON responses with HTTP status codes
- Database operation error wrapping with context
- Service-level validation before database operations

## Dependencies

```go
github.com/gin-gonic/gin v1.9.1          // HTTP framework
github.com/google/uuid v1.6.0             // UUID generation
gorm.io/gorm v1.25.5                      // ORM
gorm.io/driver/sqlite v1.5.4              // SQLite driver
```

See `go.mod` for complete dependency list.

## Future Enhancements

- [ ] Authentication/authorization (JWT tokens)
- [ ] Multiplayer sessions with real-time opponent sync
- [ ] Advanced scoring algorithms (combo multipliers)
- [ ] Daily challenges and seasonal leaderboards
- [ ] User profiles and statistics tracking
- [ ] Caching layer (Redis) for leaderboard
- [ ] GraphQL endpoint alongside REST
- [ ] WebSocket for bi-directional communication
