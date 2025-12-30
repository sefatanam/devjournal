#!/bin/bash

# DevJournal Development Script
# Runs all services with hot-reload for development
# Uses same Docker images as production for consistency

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Project root directory
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Trap to cleanup background processes on exit
cleanup() {
    echo -e "\n${YELLOW}Shutting down services...${NC}"
    kill $(jobs -p) 2>/dev/null || true
    echo -e "${GREEN}All services stopped.${NC}"
    echo -e "${YELLOW}Note: Databases are still running. Run './dev.sh stop' to stop them.${NC}"
}
trap cleanup EXIT INT TERM

# Print banner
echo -e "${BLUE}"
echo "╔═══════════════════════════════════════════════════════════╗"
echo "║                  DevJournal Development                    ║"
echo "║                    Hot-Reload Enabled                      ║"
echo "╚═══════════════════════════════════════════════════════════╝"
echo -e "${NC}"

# Check if air is installed for Go hot-reload
check_air() {
    if ! command -v air &> /dev/null; then
        echo -e "${YELLOW}Installing 'air' for Go hot-reload...${NC}"
        go install github.com/air-verse/air@latest
    fi
}

# Start databases with Docker (uses same images as production)
start_databases() {
    echo -e "${BLUE}Starting databases (PostgreSQL & MongoDB)...${NC}"
    echo -e "${YELLOW}Using production images for consistency${NC}"

    # Start only database services from production docker-compose
    docker compose -f docker/docker-compose.yml up -d postgres mongodb

    echo -e "${GREEN}Databases started!${NC}"
    echo -e "  PostgreSQL: localhost:5432"
    echo -e "  MongoDB:    localhost:27017"
    echo ""

    # Wait for databases to be ready
    echo -e "${YELLOW}Waiting for databases to be ready...${NC}"
    sleep 5
}

# Start Go API with hot-reload
start_go_api() {
    echo -e "${BLUE}Starting Go API with hot-reload (air)...${NC}"

    # Create air config if it doesn't exist
    if [ ! -f "services/go-api/.air.toml" ]; then
        cat > services/go-api/.air.toml << 'EOF'
root = "."
tmp_dir = "tmp"

[build]
  bin = "./tmp/main"
  cmd = "go build -o ./tmp/main ./cmd/api"
  delay = 1000
  exclude_dir = ["tmp", "vendor", "testdata"]
  exclude_file = []
  exclude_regex = ["_test.go"]
  exclude_unchanged = false
  follow_symlink = false
  full_bin = ""
  include_dir = []
  include_ext = ["go", "tpl", "tmpl", "html"]
  kill_delay = "2s"
  log = "build-errors.log"
  send_interrupt = false
  stop_on_error = true

[color]
  build = "yellow"
  main = "magenta"
  runner = "green"
  watcher = "cyan"

[log]
  time = false

[misc]
  clean_on_exit = true

[screen]
  clear_on_rebuild = true
EOF
    fi

    cd "$PROJECT_ROOT/services/go-api"

    # Set environment variables
    export POSTGRES_URL="postgres://devjournal:devjournal_secret@localhost:5432/devjournal?sslmode=disable"
    export MONGO_URL="mongodb://devjournal:devjournal_secret@localhost:27017"
    export MONGO_DB="devjournal"
    export JWT_SECRET="dev-secret-key"
    export HTTP_PORT="8080"
    export GRPC_PORT="8081"
    export ENVIRONMENT="development"

    air &
    GO_PID=$!

    cd "$PROJECT_ROOT"
    echo -e "${GREEN}Go API started with hot-reload!${NC}"
    echo -e "  REST API: http://localhost:8080"
    echo -e "  gRPC:     http://localhost:8081"
    echo ""
}

# Start Angular frontend with hot-reload
start_angular() {
    echo -e "${BLUE}Starting Angular frontend with hot-reload...${NC}"

    cd "$PROJECT_ROOT"
    npx nx serve web --host=0.0.0.0 &
    ANGULAR_PID=$!

    echo -e "${GREEN}Angular frontend starting...${NC}"
    echo -e "  Frontend: http://localhost:4200"
    echo ""
}

# Main execution
main() {
    cd "$PROJECT_ROOT"

    case "${1:-all}" in
        db|databases)
            start_databases
            echo -e "${GREEN}Databases are running. Press Ctrl+C to stop.${NC}"
            wait
            ;;
        api|backend)
            check_air
            start_databases
            start_go_api
            echo -e "${GREEN}Backend is running. Press Ctrl+C to stop.${NC}"
            wait
            ;;
        web|frontend)
            start_angular
            echo -e "${GREEN}Frontend is running. Press Ctrl+C to stop.${NC}"
            wait
            ;;
        all|"")
            check_air
            start_databases
            sleep 2
            start_go_api
            sleep 2
            start_angular

            echo -e "${GREEN}"
            echo "╔═══════════════════════════════════════════════════════════╗"
            echo "║              All services are running!                     ║"
            echo "╠═══════════════════════════════════════════════════════════╣"
            echo "║  Frontend:  http://localhost:4200                          ║"
            echo "║  REST API:  http://localhost:8080                          ║"
            echo "║  gRPC:      http://localhost:8081                          ║"
            echo "║  Postgres:  localhost:5432                                 ║"
            echo "║  MongoDB:   localhost:27017                                ║"
            echo "╠═══════════════════════════════════════════════════════════╣"
            echo "║  Hot-reload is enabled for Angular and Go!                 ║"
            echo "║  Press Ctrl+C to stop all services.                        ║"
            echo "╚═══════════════════════════════════════════════════════════╝"
            echo -e "${NC}"

            wait
            ;;
        stop)
            echo -e "${YELLOW}Stopping all services...${NC}"
            docker compose -f docker/docker-compose.yml stop postgres mongodb 2>/dev/null || true
            pkill -f "air" 2>/dev/null || true
            pkill -f "nx serve" 2>/dev/null || true
            echo -e "${GREEN}All services stopped.${NC}"
            ;;
        down)
            echo -e "${YELLOW}Stopping and removing all containers...${NC}"
            docker compose -f docker/docker-compose.yml down 2>/dev/null || true
            pkill -f "air" 2>/dev/null || true
            pkill -f "nx serve" 2>/dev/null || true
            echo -e "${GREEN}All services stopped and containers removed.${NC}"
            ;;
        *)
            echo "Usage: $0 [command]"
            echo ""
            echo "Commands:"
            echo "  all       Start all services with hot-reload (default)"
            echo "  db        Start only databases (PostgreSQL & MongoDB)"
            echo "  api       Start databases and Go API with hot-reload"
            echo "  web       Start only Angular frontend with hot-reload"
            echo "  stop      Stop all services (keeps data)"
            echo "  down      Stop and remove all containers (keeps data volumes)"
            ;;
    esac
}

main "$@"
