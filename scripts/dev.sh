set -e

APP_NAME="flex-service"
BUILD_DIR="tmp"
BINARY_PATH="$BUILD_DIR/$APP_NAME"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Create tmp directory
mkdir -p $BUILD_DIR

# Function to build the application
build_app() {
    echo -e "${YELLOW}Building application...${NC}"
    if go build -o $BINARY_PATH cmd/main.go; then
        echo -e "${GREEN}Build successful!${NC}"
        return 0
    else
        echo -e "${RED}Build failed!${NC}"
        return 1
    fi
}

# Function to run the application
run_app() {
    if [ -f $BINARY_PATH ]; then
        echo -e "${GREEN}Starting server...${NC}"
        ./$BINARY_PATH &
        APP_PID=$!
        echo "Server started with PID: $APP_PID"
    fi
}

# Function to kill the application
kill_app() {
    if [ ! -z "$APP_PID" ]; then
        echo -e "${YELLOW}Stopping server...${NC}"
        kill $APP_PID 2>/dev/null || true
        wait $APP_PID 2>/dev/null || true
        echo -e "${GREEN}Server stopped${NC}"
    fi
}

# Function to watch for file changes
watch_files() {
    if command -v fswatch >/dev/null 2>&1; then
        # macOS
        fswatch -o . --exclude="tmp" --exclude="bin" --exclude=".git" --exclude="coverage" | while read f; do
            echo -e "${YELLOW}File changes detected, rebuilding...${NC}"
            kill_app
            if build_app; then
                run_app
            fi
        done
    elif command -v inotifywait >/dev/null 2>&1; then
        # Linux
        while inotifywait -r -e modify,create,delete --exclude="(tmp|bin|\.git|coverage)" .; do
            echo -e "${YELLOW}File changes detected, rebuilding...${NC}"
            kill_app
            if build_app; then
                run_app
            fi
        done
    else
        echo -e "${RED}No file watcher found. Install fswatch (macOS) or inotify-tools (Linux)${NC}"
        echo -e "${YELLOW}Running without hot reload...${NC}"
        if build_app; then
            run_app
            wait $APP_PID
        fi
    fi
}

# Cleanup function
cleanup() {
    echo -e "\n${YELLOW}Cleaning up...${NC}"
    kill_app
    exit 0
}

# Set trap for cleanup
trap cleanup SIGINT SIGTERM

echo -e "${GREEN}Starting development server with hot reload...${NC}"
echo -e "${YELLOW}Press Ctrl+C to stop${NC}"

# Initial build and run
if build_app; then
    run_app
    watch_files
else
    echo -e "${RED}Initial build failed. Exiting.${NC}"
    exit 1
fi
