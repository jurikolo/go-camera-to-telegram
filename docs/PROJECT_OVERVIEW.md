# Go IP Camera Scanner & Telegram Bot - Project Overview

This document provides a comprehensive overview of the Go IP Camera Scanner project, including its structure, components, and design principles.

## Project Purpose

The Go IP Camera Scanner is an application that:

1. Discovers IP cameras on a local network (port 554)
2. Captures images from discovered cameras
3. Sends captured images to Telegram

The application is optimized for Raspberry Pi 4 with Raspbian OS.

## Project Structure

```
go-camera-to-telegram/
├── cmd/
│   └── scanner/
│       └── main.go              # Main application entry point
├── internal/
│   ├── camera/
│   │   ├── scanner.go          # Network scanning functionality
│   │   ├── rtsp.go             # RTSP client implementation
│   │   └── capture.go           # Image capture functionality
│   ├── telegram/
│   │   └── client.go           # Telegram bot client
│   └── config/
│       └── config.go          # Configuration management
├── pkg/
│   └── network/
│       └── iputil.go          # Network utility functions
├── configs/
│   └── config.yaml             # Configuration file
├── tests/
│   └── integration_test.go    # Integration tests
├── docs/
│   ├── ARCHITECTURE.md         # Application architecture
│   ├── BUILDING.md             # Build instructions
│   ├── PROJECT_STRUCTURE.md    # Project structure explanation
│   ├── PROJECT_OVERVIEW.md     # This file
│   └── SEPARATION_OF_CONCERNS.md # Design principles
├── deploy/
│   └── scanner.service         # Systemd service file
├── Makefile                    # Build automation
├── go.mod                      # Go module definition
├── go.sum                      # Go dependencies checksums
├── README.md                  # Project roadmap
├── ROADMAP.md                 # Development roadmap
└── LICENSE                    # License information
```

## Key Components

### Main Application (`cmd/scanner`)

The main application entry point that orchestrates all functionality:

- Bootstraps the application
- Loads configuration
- Initializes components
- Runs the main scanning loop

### Camera Management (`internal/camera`)

Handles all camera-related functionality:

- **Scanner**: Discovers IP cameras on the network
- **RTSP Client**: Connects to cameras using RTSP protocol
- **Connection Pool**: Manages pooled connections to RTSP cameras
- **Capture**: Extracts images from RTSP streams

### Network Scanning (`internal/network`)

Handles network-level functionality:

- **Scanner**: Scans network CIDR ranges for active hosts
- **RTSP Verification**: Verifies RTSP service availability with OPTIONS requests
- **Port Scanning**: Checks if specific ports are open on hosts

### Telegram Integration (`internal/telegram`)

Manages communication with Telegram:

- Sends images and messages to configured chat
- Handles Telegram Bot API interactions

### Configuration (`internal/config`)

Manages application configuration:

- Loads settings from YAML file
- Supports environment variable overrides
- Validates configuration values

### Network Utilities (`pkg/network`)

Provides reusable network functions:

- CIDR parsing and IP generation
- Port scanning and connectivity checks
- Local IP address detection

## Design Principles

### Separation of Concerns

The project follows Go best practices for separation of concerns:

- `cmd/`: Application entry points
- `internal/`: Application-specific business logic
- `pkg/`: Reusable libraries
- `configs/`: Configuration files
- `tests/`: Integration tests
- `docs/`: Documentation
- `deploy/`: Deployment files

### Dependency Direction

```
cmd/ -> internal/ -> pkg/
```

This ensures clean architecture with no circular dependencies.

### Encapsulation

- `internal/` packages are not accessible from outside the module
- Only exported functions are accessible outside their package
- Private implementation details are hidden

## Configuration

The application is configured through:

1. **YAML Configuration File** (`configs/config.yaml`):
   - Network settings (CIDR, timeouts, workers)
   - RTSP credentials
   - Telegram settings
   - Scan parameters

2. **Environment Variables**:
   - Override configuration file values
   - Secure credential management
   - Environment-specific settings

## Building and Deployment

### Build Process

The project includes a Makefile with targets for:

- Building for current platform
- Cross-compiling for Raspberry Pi
- Running tests
- Installing dependencies
- Cleaning build artifacts

### Deployment

The application can be deployed as:

1. **Standalone Binary**: Built and run directly
2. **Systemd Service**: For automatic startup and management
3. **Manual Execution**: For development and testing

## Testing

The project includes:

- **Unit Tests**: For individual components (to be implemented)
- **Integration Tests**: For component interactions

## Documentation

Comprehensive documentation is provided in the `docs/` directory:

- **Architecture**: Application design and components
- **Building**: Instructions for building and running
- **Project Structure**: Explanation of directory organization
- **Separation of Concerns**: Design principles and rationale

## Future Development

The project roadmap in `ROADMAP.md` outlines:

- Network scanning implementation
- RTSP camera integration
- Telegram messaging
- Testing and optimization
- Documentation and deployment

## Raspberry Pi Optimization

The application is designed with Raspberry Pi constraints in mind:

- Limited concurrent operations to prevent resource exhaustion
- Efficient memory usage
- ARM64 cross-compilation support
- Systemd service integration

## Security Considerations

- No hardcoded credentials
- Environment variable support for sensitive data
- Input validation
- Secure file permissions

## Conclusion

This project provides a solid foundation for an IP camera scanner with Telegram integration. The modular design, clear separation of concerns, and comprehensive documentation make it maintainable, testable, and extensible. The architecture supports both development and production deployment scenarios, with specific optimizations for Raspberry Pi environments.