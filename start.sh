#!/bin/bash

# DevJournal - One Command Startup Script
# Usage: ./start.sh [command]
# Commands:
#   up      - Start all services (default)
#   down    - Stop all services
#   restart - Restart all services
#   logs    - View logs
#   build   - Rebuild and start
#   clean   - Stop and remove volumes (fresh start)

set -e

DOCKER_DIR="docker"
COMPOSE_FILE="$DOCKER_DIR/docker-compose.yml"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_banner() {
    echo -e "${BLUE}"
    echo "╔═══════════════════════════════════════╗"
    echo "║         DevJournal Platform           ║"
    echo "║   Angular 21 + Go Full-Stack App      ║"
    echo "╚═══════════════════════════════════════╝"
    echo -e "${NC}"
}

print_status() {
    echo -e "${GREEN}[✓]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[!]${NC} $1"
}

print_error() {
    echo -e "${RED}[✗]${NC} $1"
}

check_docker() {
    if ! command -v docker &> /dev/null; then
        print_error "Docker is not installed. Please install Docker first."
        exit 1
    fi
    
    if ! docker info &> /dev/null; then
        print_error "Docker daemon is not running. Please start Docker."
        exit 1
    fi
    
    print_status "Docker is running"
}

start_services() {
    print_banner
    check_docker
    
    echo ""
    echo -e "${YELLOW}Starting DevJournal services...${NC}"
    echo ""
    
    docker compose -f "$COMPOSE_FILE" up -d --build
    
    echo ""
    print_status "All services started!"
    echo ""
    echo -e "${GREEN}Access the application:${NC}"
    echo "  Frontend:  http://localhost:4200"
    echo "  API:       http://localhost:8080"
    echo "  gRPC-Web:  http://localhost:9090"
    echo ""
    echo -e "${YELLOW}Database connections:${NC}"
    echo "  PostgreSQL: localhost:5432 (user: devjournal, pass: devjournal_secret)"
    echo "  MongoDB:    localhost:27017 (user: devjournal, pass: devjournal_secret)"
    echo ""
    echo "Run './start.sh logs' to view logs"
}

stop_services() {
    print_banner
    echo -e "${YELLOW}Stopping DevJournal services...${NC}"
    docker compose -f "$COMPOSE_FILE" down
    print_status "All services stopped"
}

restart_services() {
    stop_services
    start_services
}

view_logs() {
    docker compose -f "$COMPOSE_FILE" logs -f
}

build_services() {
    print_banner
    check_docker
    
    echo -e "${YELLOW}Building and starting services...${NC}"
    docker compose -f "$COMPOSE_FILE" up -d --build --force-recreate
    
    print_status "Build complete and services started!"
}

clean_start() {
    print_banner
    echo -e "${RED}Warning: This will remove all data!${NC}"
    read -p "Are you sure? (y/N): " confirm
    
    if [[ $confirm == [yY] || $confirm == [yY][eE][sS] ]]; then
        echo -e "${YELLOW}Stopping services and removing volumes...${NC}"
        docker compose -f "$COMPOSE_FILE" down -v
        print_status "Clean complete. Starting fresh..."
        start_services
    else
        echo "Cancelled."
    fi
}

show_help() {
    print_banner
    echo "Usage: ./start.sh [command]"
    echo ""
    echo "Commands:"
    echo "  up      - Start all services (default)"
    echo "  down    - Stop all services"
    echo "  restart - Restart all services"
    echo "  logs    - View logs (follow mode)"
    echo "  build   - Rebuild and start all services"
    echo "  clean   - Stop, remove volumes, and start fresh"
    echo "  help    - Show this help message"
    echo ""
}

# Main
case "${1:-up}" in
    up)
        start_services
        ;;
    down)
        stop_services
        ;;
    restart)
        restart_services
        ;;
    logs)
        view_logs
        ;;
    build)
        build_services
        ;;
    clean)
        clean_start
        ;;
    help|--help|-h)
        show_help
        ;;
    *)
        print_error "Unknown command: $1"
        show_help
        exit 1
        ;;
esac
