# HeheSwitch - IoT PlayStation Rental Management System

Always reference these instructions first and fallback to search or bash commands only when you encounter unexpected information that does not match the info here.

## Working Effectively

### Bootstrap, Build, and Test the Repository
- `go mod tidy` -- installs all dependencies. Takes ~1 second if already cached, ~10 seconds for fresh download.
- `make build` -- builds the application binary. Takes ~1 second if cached, ~30 seconds fresh. NEVER CANCEL. Set timeout to 60+ seconds.
- `go test ./...` -- runs all unit tests. Takes ~1 second if cached, ~30 seconds fresh. NEVER CANCEL. Set timeout to 60+ seconds.

### Development Workflow
- Development mode: `go run ./cmd/server` -- starts server immediately. Takes ~3 seconds to start.
- Production mode: `./dist/heheswitch` -- runs the built binary. Takes ~2 seconds to start.
- Default server port: `8080` (configurable via `PORT` environment variable)
- Database: SQLite file `heheswitch.db` (created automatically on first run)

### Code Quality and Linting
- `go fmt ./...` -- formats all Go code. Takes ~1 second.
- `go vet ./...` -- runs Go static analysis linter. Takes ~2 seconds.
- Always run both commands before committing changes.

## Validation

### Manual Testing Requirements
- **ALWAYS** test complete user scenarios after making any changes to the web application.
- **Login flow**: Test with default credentials: username `admin`, password `admin123`
- **Console management**: Test starting/stopping console rentals through the web interface
- **API endpoints**: Test key endpoints using curl with session authentication

### Complete End-to-End Validation Scenario
After any changes, ALWAYS run this complete validation:

1. **Build and start server**:
   ```bash
   make build
   PORT=8081 ./dist/heheswitch &
   sleep 3
   ```

2. **Test login and authentication**:
   ```bash
   # Login and save session cookie
   curl -s -c /tmp/cookies.txt -X POST -H "Content-Type: application/json" \
     -d '{"username":"admin","password":"admin123"}' http://localhost:8081/login
   
   # Verify login worked
   curl -s -b /tmp/cookies.txt http://localhost:8081/me
   ```

3. **Test core functionality**:
   ```bash
   # Get console status
   curl -s -b /tmp/cookies.txt http://localhost:8081/status
   
   # Test console start (should work)
   curl -s -b /tmp/cookies.txt -X POST -H "Content-Type: application/json" \
     -d '{"console_id":1,"duration_minutes":30}' http://localhost:8081/start
   ```

4. **Verify web interface**: Open browser to `http://localhost:8081` and manually test login and dashboard functionality.

**Expected Results**:
- Login should return: `{"status":"ok","user":{"id":1,"username":"admin","role":"admin",...}}`
- User info should return: `{"uid":1,"username":"admin","role":"admin",...}`
- Status should return array of 5 consoles: `[{"id":1,"name":"PS1","status":"IDLE",...},...]`
- Console start should return: `{"status":"ok"}`
- Web interface should load with login page (HTTP 200)

### Database and State
- SQLite database `heheswitch.db` is created automatically on first run
- Database includes 5 default console entries (PS1-PS5)
- Default admin user is seeded: username `admin`, password `admin123`
- To reset state: delete `heheswitch.db`, `*.db-wal`, and `*.db-shm` files

## Architecture and Code Navigation

### Project Structure
```
cmd/server/              # Main application entry point
internal/
├── adapters/           # Interface adapters (controllers, repositories)
├── api/               # HTTP handlers and routes
├── app/               # Application initialization and DI
├── config/            # Configuration management
├── domain/            # Core business logic (entities, errors, interfaces)
├── server/            # HTTP server setup
└── usecases/          # Business use cases
web/static/            # Frontend assets (HTML, CSS, JS)
```

### Key Files to Check After Changes
- **API Changes**: Always check `internal/api/handlers.go` after modifying endpoints
- **Domain Logic**: Check `internal/domain/entities/` after business logic changes
- **Database Schema**: Check `internal/db/schema.go` for database structure
- **Configuration**: Check `internal/config/config.go` for environment variables
- **Web Frontend**: Check `web/static/index.html` for UI changes

### Testing Strategy
- Unit tests exist for: domain entities, use cases, controllers, configuration
- Test files follow Go convention: `*_test.go`
- Mock implementations available in `internal/mocks/`
- Use `testify` assertions for test consistency

## Environment Configuration

### Required Environment Variables
All environment variables are optional and have sensible defaults:

- `PORT=8080` -- Server port (default: 8080)
- `DB_PATH=heheswitch.db` -- Database file path (default: heheswitch.db)
- `SQLITE_MODE=balanced` -- SQLite performance mode: aggressive|balanced|safe (default: balanced)

### Optional MQTT Configuration (IoT Integration)
- `MQTT_BROKER` -- MQTT broker URL (e.g., tcp://localhost:1883)
- `MQTT_PREFIX=ps` -- MQTT topic prefix (default: ps)
- `MQTT_USERNAME` -- MQTT authentication username
- `MQTT_PASSWORD` -- MQTT authentication password
- `MQTT_CLIENT_ID` -- MQTT client identifier

### Application Defaults
- Console count: 5 consoles (PS1-PS5)
- Default price: Rp 40,000/hour
- Admin credentials: admin/admin123 (⚠️ change in production!)

## Common Tasks

### Running Different Configurations
```bash
# Default configuration
go run ./cmd/server

# Custom port
env PORT=9090 go run ./cmd/server

# Custom database location
env DB_PATH=/tmp/test.db go run ./cmd/server

# With MQTT broker
env MQTT_BROKER=tcp://localhost:1883 go run ./cmd/server
```

### Database Operations
```bash
# View database file (if sqlite3 installed)
sqlite3 heheswitch.db ".tables"
sqlite3 heheswitch.db "SELECT * FROM consoles;"

# Reset database (delete files)
rm -f heheswitch.db heheswitch.db-*
```

### Build Variations
```bash
# Local build (current OS/architecture)
make build

# Linux ARM64 build (for Raspberry Pi)
make build-arm64

# Manual build with custom flags
CGO_ENABLED=0 go build -o dist/heheswitch ./cmd/server
```

## API Endpoints Reference

### Authentication (Public)
- `POST /login` -- User login with username/password
- `GET /login` -- Login page HTML

### Console Management (Authenticated)
- `GET /status` -- Real-time status of all consoles
- `POST /start` -- Start console rental
- `POST /extend` -- Extend rental duration  
- `POST /stop` -- Stop console rental
- `GET /transactions/:console_id` -- Rental history

### User Management (Admin Only)
- `GET /users` -- List all users
- `POST /users` -- Create new user
- `DELETE /users/:id` -- Delete user

### WebSocket
- `GET /ws` -- WebSocket connection for real-time updates

## Troubleshooting

### Common Issues and Solutions

| Problem | Solution |
|---------|----------|
| Port already in use | Change `PORT` environment variable: `PORT=8081 go run ./cmd/server` |
| Database locked | Stop application, delete `*.db-wal` and `*.db-shm` files |
| Build fails | Run `go mod tidy` to ensure dependencies are correct |
| Tests fail | Check if any test database files exist and remove them |
| Web assets not loading | Ensure `web/static/` directory exists with HTML/CSS/JS files |

### Debugging Commands
```bash
# Check if server is running and responsive
curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/login

# Test complete login flow
curl -s -c /tmp/test_cookies.txt -X POST -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}' http://localhost:8080/login

# Check authentication works
curl -s -b /tmp/test_cookies.txt http://localhost:8080/me

# View server logs in development
go run ./cmd/server

# Check binary exists and size after build
ls -la dist/heheswitch
file dist/heheswitch  # Should show "ELF 64-bit LSB executable"

# Verify database was created
ls -la *.db*  # Should show heheswitch.db and possibly .wal/.shm files
```

### Dependencies and Version Requirements
- **Go**: Version 1.24+ required (check with `go version`)
- **Operating System**: Linux, macOS, or Windows supported
- **Memory**: Minimum 64MB RAM
- **Storage**: Minimum 50MB free space
- **Network**: HTTP port access (default 8080)

## Additional Notes

### Clean Architecture Pattern
This codebase follows clean architecture principles:
- **Domain** layer contains business entities and rules
- **Use Cases** layer contains application business logic
- **Adapters** layer contains interface implementations
- **Infrastructure** layer contains external dependencies

### Web Technology Stack
- **Backend**: Go with Fiber web framework
- **Database**: SQLite with embedded storage
- **Frontend**: Vanilla HTML, CSS, JavaScript with WebSocket
- **Real-time**: WebSocket for live console status updates
- **IoT**: MQTT protocol support for device control

### Development Best Practices
- Always run tests before committing: `go test ./...`
- Format code before committing: `go fmt ./...`
- Use `go vet ./...` to catch common mistakes
- Test web functionality manually after UI changes
- Validate all API endpoints after backend changes
- Check database schema consistency after domain changes