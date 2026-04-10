#!/bin/bash

# Ravel Comprehensive Test Runner
# Runs all tests with various configurations and generates reports

set -e

COLOR_RED='\033[0;31m'
COLOR_GREEN='\033[0;32m'
COLOR_YELLOW='\033[1;33m'
COLOR_BLUE='\033[0;34m'
COLOR_RESET='\033[0m'

log_info() {
    echo -e "${COLOR_BLUE}[INFO]${COLOR_RESET} $1"
}

log_success() {
    echo -e "${COLOR_GREEN}[SUCCESS]${COLOR_RESET} $1"
}

log_warning() {
    echo -e "${COLOR_YELLOW}[WARNING]${COLOR_RESET} $1"
}

log_error() {
    echo -e "${COLOR_RED}[ERROR]${COLOR_RESET} $1"
}

COVERAGE_DIR="coverage"
TEST_TIMEOUT="10m"
VERBOSE=${VERBOSE:-""}

echo "======================================"
echo "  Ravel Test Runner"
echo "======================================"
echo ""

# Parse arguments
RUN_UNIT=true
RUN_INTEGRATION=false
RUN_COVERAGE=false
RUN_RACE=false
RUN_SHORT=false

while [[ $# -gt 0 ]]; do
    case $1 in
        --integration)
            RUN_INTEGRATION=true
            shift
            ;;
        --coverage)
            RUN_COVERAGE=true
            shift
            ;;
        --race)
            RUN_RACE=true
            shift
            ;;
        --short)
            RUN_SHORT=true
            shift
            ;;
        --all)
            RUN_UNIT=true
            RUN_INTEGRATION=true
            RUN_COVERAGE=true
            RUN_RACE=true
            shift
            ;;
        --help)
            echo "Usage: $0 [OPTIONS]"
            echo ""
            echo "Options:"
            echo "  --integration  Run integration tests"
            echo "  --coverage     Generate coverage report"
            echo "  --race         Run with race detector"
            echo "  --short        Run only short tests"
            echo "  --all          Run all test types"
            echo "  --help         Show this help message"
            exit 0
            ;;
        *)
            log_error "Unknown option: $1"
            exit 1
            ;;
    esac
done

# Prepare coverage directory
if [ "$RUN_COVERAGE" = true ]; then
    mkdir -p $COVERAGE_DIR
    rm -f $COVERAGE_DIR/*.out
fi

# Function to run tests for a package
run_package_tests() {
    local package=$1
    local test_flags=$2
    local description=$3

    log_info "Testing $package - $description"

    if go test $test_flags -timeout $TEST_TIMEOUT $package; then
        log_success "✓ $package passed"
        return 0
    else
        log_error "✗ $package failed"
        return 1
    fi
}

FAILED_TESTS=()
PASSED_TESTS=0
TOTAL_TESTS=0

# 1. Run unit tests
if [ "$RUN_UNIT" = true ]; then
    log_info "Running unit tests..."
    echo ""

    TEST_FLAGS="-v"
    if [ "$RUN_SHORT" = true ]; then
        TEST_FLAGS="$TEST_FLAGS -short"
    fi
    if [ "$RUN_RACE" = true ]; then
        TEST_FLAGS="$TEST_FLAGS -race"
    fi

    # Test specific packages
    PACKAGES=(
        "./api/..."
        "./core/..."
        "./ravel/..."
        "./runtime/..."
        "./agent/..."
        "./pkg/..."
    )

    for pkg in "${PACKAGES[@]}"; do
        ((TOTAL_TESTS++))
        if run_package_tests "$pkg" "$TEST_FLAGS" "unit tests"; then
            ((PASSED_TESTS++))
        else
            FAILED_TESTS+=("$pkg")
        fi
        echo ""
    done
fi

# 2. Run integration tests
if [ "$RUN_INTEGRATION" = true ]; then
    log_info "Running integration tests..."
    echo ""

    if [ -d "tests/integration" ]; then
        ((TOTAL_TESTS++))
        if run_package_tests "./tests/integration/..." "-v -tags=integration" "integration tests"; then
            ((PASSED_TESTS++))
        else
            FAILED_TESTS+=("integration")
        fi
        echo ""
    else
        log_warning "No integration tests directory found"
    fi
fi

# 3. Generate coverage report
if [ "$RUN_COVERAGE" = true ]; then
    log_info "Generating coverage report..."
    echo ""

    go test -coverprofile=$COVERAGE_DIR/coverage.out -covermode=atomic ./... 2>&1 | tee $COVERAGE_DIR/test.log

    # Generate HTML report
    go tool cover -html=$COVERAGE_DIR/coverage.out -o $COVERAGE_DIR/coverage.html

    # Generate text summary
    go tool cover -func=$COVERAGE_DIR/coverage.out > $COVERAGE_DIR/coverage.txt

    # Calculate total coverage
    TOTAL_COVERAGE=$(go tool cover -func=$COVERAGE_DIR/coverage.out | grep total | awk '{print $3}')

    log_success "Coverage report generated:"
    echo "  - HTML: $COVERAGE_DIR/coverage.html"
    echo "  - Text: $COVERAGE_DIR/coverage.txt"
    echo "  - Total coverage: $TOTAL_COVERAGE"
    echo ""
fi

# 4. Run go vet
log_info "Running go vet..."
if go vet ./...; then
    log_success "✓ go vet passed"
else
    log_error "✗ go vet found issues"
    FAILED_TESTS+=("go vet")
fi
echo ""

# 5. Run staticcheck if available
if command -v staticcheck &> /dev/null; then
    log_info "Running staticcheck..."
    if staticcheck ./...; then
        log_success "✓ staticcheck passed"
    else
        log_error "✗ staticcheck found issues"
        FAILED_TESTS+=("staticcheck")
    fi
    echo ""
else
    log_warning "staticcheck not installed, skipping"
fi

# Summary
echo "======================================"
echo "  Test Summary"
echo "======================================"
echo ""

if [ $TOTAL_TESTS -gt 0 ]; then
    echo "Tests run: $TOTAL_TESTS"
    echo "Passed: $PASSED_TESTS"
    echo "Failed: $((TOTAL_TESTS - PASSED_TESTS))"
    echo ""
fi

if [ ${#FAILED_TESTS[@]} -eq 0 ]; then
    log_success "All tests passed! 🎉"
    exit 0
else
    log_error "Some tests failed:"
    for test in "${FAILED_TESTS[@]}"; do
        echo "  - $test"
    done
    exit 1
fi
