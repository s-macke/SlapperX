This file provides guidance when working with code in this repository.

## Project Overview

SlapperX is a Go-based HTTP load testing tool that sends configurable HTTP requests and displays real-time performance metrics in a terminal histogram interface. It supports .http file format for request definitions and provides interactive rate control.

## Build and Development Commands

```bash
# Build the project (creates binary in current directory)
go build

# Run unit tests
go test ./src/...

# Run the tool
./slapperx -targets targets.http -workers 8 -rate 50

```

## Architecture Overview

**Core Components:**
- **slapper.go**: Main orchestrator that initializes and coordinates all components
- **targeter.go**: Load generation engine managing worker goroutines and HTTP execution
- **ticker.go**: Request rate controller with dynamic rate adjustment support
- **stats.go**: Thread-safe statistics collection using atomic counters
- **movingWindow.go**: Time-slotted response time histogram data with automatic rotation
- **ui.go**: Real-time terminal visualization of performance metrics
- **rampupController.go**: Gradual rate increases and runtime rate management
- **keyboard.go + input.go**: Interactive controls for rate adjustment and commands

**Communication Patterns:**
- Channel-based architecture for thread-safe component communication
- Worker pool pattern with shared tick channel for load distribution
- Producer-consumer model: ticker → workers → stats → UI
- Observer pattern for real-time metrics updates
- Errors are created using the `errors` package for consistent error handling

**Key Data Flow:**
1. Ticker generates timing events at specified rate
2. Worker pool executes HTTP requests round-robin from parsed .http requests
3. Workers update global atomic counters and feed timing data to MovingWindow
4. UI aggregates data every 200ms for histogram display
5. Keyboard input drives rate changes via RampUpController

**Tests:**
- Unit tests always test all code paths, including error handling
- test http requests must be defined in a testdata directory

**HTTP Processing:**
- Custom HTTP client with connection tracking (`tracing/tracingClient.go`)
- Request parsing from .http files (`httpfile/` package)
- Round-robin request selection using atomic counters
- Time-slotted response data collection for continuous histogram updates

## Request Format

Uses JetBrains .http file format with ### separators between requests. Supports headers, request bodies, and multiple HTTP methods.