#!/bin/bash

# scripts/test.sh - Enhanced test runner script

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
COVERAGE_THRESHOLD=70
TEST_TIMEOUT=30s

echo -e "${BLUE}🧪 Go Clean Gin API - Test Runner${NC}"
echo "======================================"

# Function to print colored output
print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to check if Go is available
check_go() {
    if ! command -v go &> /dev/null; then
        print_error "Go is not installed or not in PATH"
        exit 1
    fi
    print_status "Go version: $(go version)"
}

# Function to run unit tests
run_unit_tests() {
    print_status "Running unit tests..."
    
    if go test -v -short -timeout ${TEST_TIMEOUT} ./internal/...; then
        print_status "✅ Unit tests passed"
    else
        print_error "❌ Unit tests failed"
        exit 1
    fi
}

# Function to run all tests with coverage
run_tests_with_coverage() {
    print_status "Running all tests with coverage..."
    
    # Create coverage directory if it doesn't exist
    mkdir -p coverage
    
    # Run tests with coverage
    if go test -v -race -coverprofile=coverage/coverage.out -timeout ${TEST_TIMEOUT} ./...; then
        print_status "✅ All tests passed"
    else
        print_error "❌ Some tests failed"
        exit 1
    fi
    
    # Generate coverage report
    if [ -f coverage/coverage.out ]; then
        print_status "Generating coverage reports..."
        
        # Generate HTML report
        go tool cover -html=coverage/coverage.out -o coverage/coverage.html
        
        # Generate function coverage report
        go tool cover -func=coverage/coverage.out > coverage/coverage.txt
        
        # Calculate total coverage percentage
        COVERAGE=$(go tool cover -func=coverage/coverage.out | grep total | awk '{print $3}' | sed 's/%//')
        
        if [ ! -z "$COVERAGE" ]; then
            print_status "Total coverage: ${COVERAGE}%"
            
            # Check coverage threshold
            if (( $(echo "$COVERAGE >= $COVERAGE_THRESHOLD" | bc -l) )); then
                print_status "✅ Coverage threshold met (${COVERAGE_THRESHOLD}%)"
            else
                print_warning "⚠️ Coverage below threshold: ${COVERAGE}% < ${COVERAGE_THRESHOLD}%"
            fi
        else
            print_warning "Could not determine coverage percentage"
        fi
        
        print_status "Coverage reports generated:"
        print_status "  - HTML: coverage/coverage.html"
        print_status "  - Text: coverage/coverage.txt"
    else
        print_warning "Coverage file not found"
    fi
}

# Function to run integration tests
run_integration_tests() {
    print_status "Running integration tests..."
    
    if [ -f test/integration_test.go ]; then
        if go test -v -tags=integration -timeout ${TEST_TIMEOUT} ./test/...; then
            print_status "✅ Integration tests passed"
        else
            print_error "❌ Integration tests failed"
            exit 1
        fi
    else
        print_warning "No integration tests found"
    fi
}

# Function to run benchmark tests
run_benchmarks() {
    print_status "Running benchmark tests..."
    
    if go test -bench=. -benchmem -timeout ${TEST_TIMEOUT} ./...; then
        print_status "✅ Benchmark tests completed"
    else
        print_warning "⚠️ Benchmark tests encountered issues"
    fi
}

# Function to check test files
check_test_files() {
    print_status "Checking test files..."
    
    # Find Go files without corresponding test files
    MISSING_TESTS=""
    
    for go_file in $(find ./internal -name "*.go" -not -name "*_test.go"); do
        test_file="${go_file%.*}_test.go"
        if [ ! -f "$test_file" ]; then
            MISSING_TESTS="$MISSING_TESTS\n  - $go_file"
        fi
    done
    
    if [ ! -z "$MISSING_TESTS" ]; then
        print_warning "Files without test files:$MISSING_TESTS"
    else
        print_status "✅ All files have corresponding test files"
    fi
}

# Function to run race detection
run_race_tests() {
    print_status "Running tests with race detection..."
    
    if go test -race -short -timeout ${TEST_TIMEOUT} ./...; then
        print_status "✅ No race conditions detected"
    else
        print_error "❌ Race conditions detected"
        exit 1
    fi
}

# Function to clean up test artifacts
cleanup() {
    print_status "Cleaning up test artifacts..."
    
    # Remove temporary test files
    find . -name "*.test" -delete 2>/dev/null || true
    find . -name "*.prof" -delete 2>/dev/null || true
    
    print_status "✅ Cleanup completed"
}

# Function to validate test environment
validate_environment() {
    print_status "Validating test environment..."
    
    # Check if required environment variables are set for integration tests
    if [ "$RUN_INTEGRATION" = "true" ]; then
        if [ -z "$DB_HOST" ] || [ -z "$DB_PORT" ]; then
            print_warning "Database environment variables not set for integration tests"
            print_warning "Set DB_HOST and DB_PORT or disable integration tests"
        fi
    fi
    
    # Check Go module
    if [ ! -f go.mod ]; then
        print_error "go.mod not found. Run this script from the project root"
        exit 1
    fi
    
    print_status "✅ Environment validation passed"
}

# Function to show usage
show_usage() {
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  -u, --unit          Run unit tests only"
    echo "  -i, --integration   Run integration tests only"
    echo "  -b, --benchmark     Run benchmark tests"
    echo "  -r, --race          Run race detection tests"
    echo "  -c, --coverage      Run tests with coverage (default)"
    echo "  -a, --all           Run all types of tests"
    echo "  -v, --verbose       Verbose output"
    echo "  -h, --help          Show this help message"
    echo ""
    echo "Environment variables:"
    echo "  COVERAGE_THRESHOLD  Coverage threshold percentage (default: 70)"
    echo "  TEST_TIMEOUT        Test timeout duration (default: 30s)"
    echo "  RUN_INTEGRATION     Enable integration tests (true/false)"
}

# Main execution
main() {
    local run_unit=false
    local run_integration=false
    local run_benchmarks=false
    local run_race=false
    local run_coverage=true
    local run_all=false
    local verbose=false
    
    # Parse command line arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            -u|--unit)
                run_unit=true
                run_coverage=false
                shift
                ;;
            -i|--integration)
                run_integration=true
                run_coverage=false
                shift
                ;;
            -b|--benchmark)
                run_benchmarks=true
                run_coverage=false
                shift
                ;;
            -r|--race)
                run_race=true
                run_coverage=false
                shift
                ;;
            -c|--coverage)
                run_coverage=true
                shift
                ;;
            -a|--all)
                run_all=true
                shift
                ;;
            -v|--verbose)
                verbose=true
                shift
                ;;
            -h|--help)
                show_usage
                exit 0
                ;;
            *)
                print_error "Unknown option: $1"
                show_usage
                exit 1
                ;;
        esac
    done
    
    # Check prerequisites
    check_go
    validate_environment
    
    # Set trap for cleanup on exit
    trap cleanup EXIT
    
    # Run tests based on options
    if [ "$run_all" = true ]; then
        run_unit_tests
        run_tests_with_coverage
        run_integration_tests
        run_benchmarks
        run_race_tests
    elif [ "$run_unit" = true ]; then
        run_unit_tests
    elif [ "$run_integration" = true ]; then
        run_integration_tests
    elif [ "$run_benchmarks" = true ]; then
        run_benchmarks
    elif [ "$run_race" = true ]; then
        run_race_tests
    elif [ "$run_coverage" = true ]; then
        run_tests_with_coverage
    fi
    
    # Additional checks
    check_test_files
    
    echo ""
    print_status "🎉 Test execution completed successfully!"
    
    # Show summary
    if [ -f coverage/coverage.out ]; then
        echo ""
        echo "📊 Test Summary:"
        echo "=================="
        if [ ! -z "$COVERAGE" ]; then
            echo "Total Coverage: ${COVERAGE}%"
        fi
        echo "Coverage Report: coverage/coverage.html"
        echo "Coverage Details: coverage/coverage.txt"
    fi
}

# Execute main function with all arguments
main "$@"