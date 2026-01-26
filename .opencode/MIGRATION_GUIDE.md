# Migration Guide: Understanding the Go Backend

## For Developers Familiar with NestJS

This guide helps you understand the Go backend if you're coming from a NestJS background.

### Mental Model Shift

| Concept                  | NestJS                 | Go                   | How to Think About It            |
| ------------------------ | ---------------------- | -------------------- | -------------------------------- |
| **Modules**              | @Module decorators     | Package organization | Files in same directory = module |
| **Services**             | @Injectable classes    | Struct + methods     | Objects with behavior            |
| **Dependency Injection** | Constructor injection  | Pass as parameters   | Explicit > Implicit              |
| **Controllers**          | @Controller classes    | Handler functions    | Route → Function                 |
| **Middleware**           | @Middleware decorators | Gin middleware       | Function → Function → Function   |
| **Exceptions**           | throw HttpException    | return error         | Go errors are values             |
| **Database**             | TypeORM decorators     | GORM methods         | Code talks to DB directly        |
| **Testing**              | TestingModule          | Test database        | Setup → Test → Teardown          |

### Key Differences

#### 1. No Decorators

**NestJS:**

```typescript
@Controller('api/game')
export class GameController {
  @Post('start')
  @UseGuards(AuthGuard)
  async startGame(@Body() req: StartGameDto) {
    return this.gameService.startGame(req);
  }
}
```

**Go:**

```go
func (gr *GameRoutes) StartGame(c *gin.Context) {
  var req StartGameRequest
  if err := c.ShouldBindJSON(&req); err != nil {
    c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
    return
  }
  // ... handler code
}

// Register in main.go
game.POST("/start", gr.StartGame)
```

#### 2. Explicit Error Handling

**NestJS:**

```typescript
try {
  await this.sessionService.create();
} catch (error) {
  throw new HttpException('Failed', HttpStatus.INTERNAL_SERVER_ERROR);
}
```

**Go:**

```go
session, err := s.sessionService.Create()
if err != nil {
  return nil, fmt.Errorf("failed to create session: %w", err)
}
```

#### 3. Dependency Injection

**NestJS (Automatic):**

```typescript
@Injectable()
export class GameService {
  constructor(
    private sessionService: SessionService,
    private wordService: WordService,
  ) {}
}
```

**Go (Explicit):**

```go
type GameService struct {
  sessionService *SessionService
  wordService    *WordService
}

func NewGameService(ss *SessionService, ws *WordService) *GameService {
  return &GameService{
    sessionService: ss,
    wordService:    ws,
  }
}
```

#### 4. Concurrency

**NestJS (Event-based):**

```typescript
@EventListener('timer.tick')
async onTimerTick(event: TimerTickEvent) {
  // Handle tick
}

this.eventEmitter.emit('timer.tick', { sessionId });
```

**Go (Goroutine-based):**

```go
go func() {
  ticker := time.NewTicker(1 * time.Second)
  defer ticker.Stop()

  for range ticker.C {
    // Handle tick
  }
}()
```

### File Organization

#### NestJS Structure

```
src/
├── main.ts                          # Bootstrap
├── app.module.ts                    # Root module
├── database/
│   └── database.provider.ts         # DB config
├── entities/
│   ├── session.entity.ts            # @Entity decorator
│   └── result.entity.ts             # @Entity decorator
├── services/
│   ├── game/
│   │   ├── game.service.ts          # @Injectable
│   │   ├── game.module.ts           # @Module
│   │   └── game.spec.ts             # Tests
│   └── session/
│       └── ...
└── controllers/
    └── game.controller.ts           # @Controller
```

#### Go Equivalent

```
internal/
├── models/
│   └── models.go                    # Struct definitions
├── services/
│   ├── game.service.go              # Service struct + methods
│   ├── game_test.go                 # Tests (same file)
│   ├── session.service.go
│   └── session_test.go
├── routes/
│   ├── game_routes.go               # Handler registration
│   └── game_routes_test.go          # Route tests
├── database/
│   └── database.go                  # DB initialization
└── middleware/
    └── cors.go                      # Middleware functions
cmd/
└── main.go                          # Bootstrap

```

### Data Flow Comparison

#### Creating a Session

**NestJS:**

```
Request → Controller.startGame()
  ↓
Constructor inject GameService
  ↓
GameService.startGame()
  ↓
Constructor inject SessionService
  ↓
SessionService.create()
  ↓
TypeORM Repository.save()
  ↓
Response
```

**Go:**

```
Request → GameRoutes.StartGame(c *gin.Context)
  ↓
Parse JSON with c.ShouldBindJSON()
  ↓
Call s.sessionService.Create()  [already injected]
  ↓
SessionService.Create()
  ↓
s.db.Create(session)  [already has GORM instance]
  ↓
c.JSON(http.StatusCreated, response)
```

### Service Method Patterns

#### Session Service Example

**NestJS:**

```typescript
@Injectable()
export class SessionService {
  constructor(private repository: Repository<SessionEntity>) {}

  async create(autoVoice: boolean): Promise<SessionEntity> {
    const session = this.repository.create({
      id: uuidv4(),
      autoVoice,
      // ...
    });
    return this.repository.save(session);
  }

  async findById(id: string): Promise<SessionEntity | null> {
    return this.repository.findOne({ where: { id } });
  }
}
```

**Go:**

```go
type SessionService struct {
  db *gorm.DB
}

func (s *SessionService) Create(autoVoice bool) (*SessionEntity, error) {
  session := &SessionEntity{
    ID:        uuid.New().String(),
    AutoVoice: autoVoice,
    // ...
  }
  if err := s.db.Create(session).Error; err != nil {
    return nil, fmt.Errorf("failed to create: %w", err)
  }
  return session, nil
}

func (s *SessionService) FindByID(id string) (*SessionEntity, error) {
  var session SessionEntity
  if err := s.db.First(&session, "id = ?", id).Error; err != nil {
    return nil, fmt.Errorf("failed to find: %w", err)
  }
  return &session, nil
}
```

### Testing Pattern Comparison

#### NestJS Testing

```typescript
describe('GameService', () => {
  let service: GameService;
  let module: TestingModule;

  beforeEach(async () => {
    module = await Test.createTestingModule({
      providers: [GameService, SessionService],
    }).compile();

    service = module.get<GameService>(GameService);
  });

  it('should start game', async () => {
    const result = await service.startGame('session-id');
    expect(result.phase).toBe('playing');
  });
});
```

#### Go Testing

```go
func TestGameService_StartGame(t *testing.T) {
  db := setupTestDB(t)
  ws := setupTestWordService(t)
  ss := NewSessionService(db)
  ls := NewLeaderboardService(db)
  gs := NewGameService(ss, ws, ls)

  session, _ := ss.Create(false)
  started, err := gs.StartGame(session.ID)

  if err != nil {
    t.Fatalf("Failed: %v", err)
  }
  if started.Phase != "playing" {
    t.Errorf("Expected phase playing, got %s", started.Phase)
  }
}
```

### API Routes

#### NestJS Controller

```typescript
@Controller('api/game')
export class GameController {
  constructor(private gameService: GameService) {}

  @Post('start')
  async startGame(@Body() req: StartGameDto) {
    return this.gameService.startGame(req);
  }

  @Post('answer')
  async submitAnswer(@Body() req: SubmitAnswerDto) {
    return this.gameService.submitAnswer(req);
  }

  @Get('sse/:sessionId')
  async sse(@Param('sessionId') id: string) {
    return this.gameService.sse(id);
  }
}
```

#### Go Routes

```go
func (gr *GameRoutes) Register(r *gin.Engine) {
  api := r.Group("/api")
  game := api.Group("/game")

  game.POST("/start", gr.StartGame)
  game.POST("/answer", gr.SubmitAnswer)
  game.GET("/sse/:sessionId", gr.SSE)
}
```

### Environment & Configuration

#### NestJS (.env)

```env
NODE_ENV=development
PORT=3000
DATABASE_URL=sqlite:./atoyr.db
```

Used via `ConfigService` injected into services.

#### Go

```bash
export NODE_ENV=development
export PORT=3000
export DATABASE_PATH=./atoyr.db
```

Accessed via `os.Getenv()` in functions.

### Common Gotchas

#### 1. Pointer vs Value Receivers

Go method receivers can be pointers or values:

```go
// Pointer receiver (mutates the receiver)
func (s *SessionService) Update(session *SessionEntity) error {
  return s.db.Save(session).Error
}

// Value receiver (doesn't mutate)
func (w WordService) GetRandomWord() string {
  // ...
}
```

For most services, use pointer receivers so they can access the injected dependencies.

#### 2. Error Handling

Go treats errors as return values, not exceptions:

```go
// Do this
result, err := service.DoSomething()
if err != nil {
  return nil, fmt.Errorf("context: %w", err)
}

// Not this (will panic)
result := service.MustDoSomething()
```

#### 3. Database Query Syntax

GORM queries are chainable:

```go
// vs TypeORM
this.db.find({ where: { id: 'id' } })

// GORM
s.db.First(&session, "id = ?", id)
s.db.Where("score > ?", 100).Find(&sessions)
s.db.Order("score DESC").Limit(10).Find(&entries)
```

#### 4. JSON Struct Tags

Go requires struct tags for JSON serialization:

```go
type User struct {
  ID   string `json:"id"`              // Export as "id" in JSON
  Name string `json:"name"`
  Age  int    `json:"age,omitempty"`   // Omit if zero-value
}
```

#### 5. Interface Casting

Go doesn't auto-convert types:

```go
// NestJS (automatic type conversion)
async process(id: string) {
  // id is a string
}

// Go (explicit)
result, ok := value.(string)
if !ok {
  return errors.New("not a string")
}
```

### Debugging

#### NestJS

```bash
npm run dev:debug
# Chrome DevTools debugging
```

#### Go

```bash
# Using Delve debugger
dlv debug ./cmd
(dlv) break main.main
(dlv) continue

# Or print debugging
log.Printf("Session: %+v", session)
```

### Performance Differences

Go implementation will be **faster**:

- **Startup**: NestJS ~2s, Go ~50ms (40x)
- **Memory**: NestJS ~120MB idle, Go ~15MB (8x)
- **Throughput**: Go ~5000 req/s vs NestJS ~1000 req/s (5x)

This is due to compiled binary vs interpreted TypeScript.

### Migration Checklist

When comparing implementations:

- [ ] Services have identical method signatures
- [ ] Database operations return same data
- [ ] HTTP responses match JSON structure
- [ ] Error messages are descriptive
- [ ] Tests cover same scenarios
- [ ] Middleware provides same functionality
- [ ] SSE events trigger at same times
- [ ] Leaderboard queries return identical results

### Next Steps

1. **Review** `internal/services/game.service.go` - Core game logic
2. **Study** `cmd/main.go` - Application bootstrap
3. **Test** `go test -v ./...` - See tests in action
4. **Debug** `NODE_ENV=development go run ./cmd` - See logs
5. **Explore** API endpoints with curl
6. **Deploy** - Read deployment section in README

### Resources

- [Go Basics](https://golang.org/doc/effective_go)
- [Gin Web Framework](https://github.com/gin-gonic/gin)
- [GORM Documentation](https://gorm.io)
- [Server README](packages/server/README.md)

You've got this! Go is simpler than it looks! 🚀
