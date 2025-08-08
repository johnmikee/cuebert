# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build and Development Commands

### Building
- `make build` or `make cuebert` - Build for current platform
- `make dev` - Build development version with version info
- `make xp-cuebert` - Cross-platform build (Darwin and Linux)
- `make release-zip` - Create release zip with cross-platform builds

### Testing
- `go test -cover ./...` - Run all tests with coverage
- `go test ./path/to/package` - Run tests for specific package

### Code Quality
- `make lint` - Run golangci-lint with project configuration
- `make fmt` - Format code with gofumpt
- `make vet` - Run go vet
- `make static-check` - Run staticcheck
- `make tidy` - Clean up go.mod dependencies

### Database Setup
- `make run-docker-postgres` - Start PostgreSQL in Docker for local development
- `make clean-postgres-db` - Stop and remove PostgreSQL container
- Database migrations are handled automatically on startup

### Running Cuebert
- `./build/darwin/cuebert` - Run with default settings (testing mode, dev environment)
- Key flags:
  - `-testing=false` - Disable testing mode for production
  - `-env-type=prod` - Set to production environment
  - `-init=false` - Skip initialization phase and start immediately
  - `-log-level=debug` - Set logging level (trace, debug, info, warn, error)

## Architecture Overview

### Core Components

**Main Entry Point** (`cuebert/cuebert.go`)
- Orchestrates bot lifecycle, database connections, and task management
- Monitors for start/stop/reload signals
- Manages goroutines for bot responses, task diffing, and routine checks

**Bot Module** (`bot/`)
- Handles Slack interactions via socket mode
- Processes commands, modals, and user messages
- Manages reminder scheduling and user notifications

**Task System** (`cuebert/task/`)
- **Compliance**: Tracks compliance tasks and user completion status
- **Update**: Manages OS update tracking and device compliance
- **Shared**: Common functionality for blocks, messaging, and user management

**Database Layer** (`db/`)
- PostgreSQL database with automatic migrations
- Tables: users, devices, bot_results, compliance_tracking, exclusions
- Connection pooling via pgx/v5

**Integration Clients**
- **IDP** (`idp/`): Okta integration for user/group management
- **MDM** (`mdm/`): Jamf/Kandji integration for device management
- **Slack** (`action/slack/`): Slack-specific actions and group DMs

### Key Workflows

1. **Initialization**: Load config → Connect to DB → Run migrations → Populate tables → Start bot
2. **Routine Tasks**:
   - **Diff** (60 min): Sync data between MDM/IDP and database
   - **Check** (30 min): Send messages to users based on compliance status
   - **Poll** (15 min): Check for pending custom reminders
3. **User Interaction**: Bot responds to commands/modals → Updates database → Schedules reminders

### Configuration

Configuration is loaded in priority order:
1. Environment variables (prefix: `CUEBERT_`)
2. System keychain (dev mode)
3. JSON config file

### Database Schema

The system uses PostgreSQL with these main tables:
- `users`: Slack user information and manager relationships
- `devices`: Device inventory from MDM
- `bot_results`: Interaction tracking for OS updates
- `compliance_tracking`: Task completion tracking
- `exclusions`: Users/devices excluded from notifications

### Testing Approach

- Unit tests using standard Go testing package with testify assertions
- Mock interfaces for external services (Slack, MDM, IDP)
- Test database migrations in `db/migrations/test/`
- Use `-testing=true` flag for dry-run mode in production environment