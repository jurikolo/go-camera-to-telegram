# Project Structure & Separation of Concerns

This document explains the Go project structure for the IP Camera Scanner with Telegram Integration, following Go best practices for organizing code.

## Project Structure

```
go-camera-to-telegram/
├── cmd/
│   └── scanner/
│       └── main.go              # Main application entry point
├── internal/
│   ├── camera/
│   │   ├── scanner.go          # Network scanning functionality
│   │   ├── rtsp.go              # RTSP client implementation
│   │   └── capture.go            # Image capture functionality
│   ├── network/
│   │   └── scanner.go            # Network scanning and RTSP verification
│   ├── telegram/
│   │   └── client.go            # Telegram bot client
│   └── config/
│       └── config.go             # Configuration management
├── pkg/
│   └── network/
│       └── iputil.go             # Network utility functions
├── configs/
│   └── config.yaml               # Configuration file
├── tests/                        # Test files
├── docs/                         # Documentation
├── Makefile                      # Build automation
├── README.md                    # Project overview and usage
├── ROADMAP.md                    # Development roadmap
└── LICENSE                       # License information
```

## Separation of Concerns

This project follows the principles of separation of concerns by organizing code into distinct packages with specific responsibilities:

### `cmd/` Directory

The `cmd/` directory contains the main application entry points. Each subdirectory corresponds to a different executable:

- **`cmd/scanner/`**: Contains `main.go` which is the entry point for the IP camera scanner application.

The main function in `cmd/scanner/main.go` is responsible for:
- Bootstrapping the application
- Parsing command-line arguments
- Initializing dependencies
- Starting the main application loop

This follows the Go best practice of keeping the `main` package minimal and focused only on application startup.

### `internal/` Directory

The `internal/` directory contains packages that are only accessible within this module. These packages contain business logic that is specific to this application and should not be used by external projects.

#### `internal/camera/`

This package handles all functionality related to IP camera discovery, RTSP connections, and image capture:

- **`scanner.go`**: Implements network scanning for IP cameras
- **`rtsp.go`**: Handles RTSP protocol communication with cameras
- **`capture.go`**: Manages image capture from RTSP streams

#### `internal/network/`

This package handles network-level functionality:

- **`scanner.go`**: Implements network scanning for active hosts and RTSP service verification

#### `internal/telegram/`

This package handles all Telegram bot integration:

- **`client.go`**: Implements the Telegram Bot API client and message sending functionality

#### `internal/config/`

This package handles configuration management:

- **`config.go`**: Loads and manages application configuration from files and environment variables

### `pkg/` Directory

The `pkg/` directory contains packages that are intended to be reusable libraries that could potentially be used by other projects. These packages should contain generic functionality that is not specific to this application.

#### `pkg/network/`

This package contains generic network utility functions:

- **`iputil.go`**: Provides IP address manipulation and network scanning utilities

The distinction between `internal/` and `pkg/` is important:
- `internal/` packages are application-specific and may change frequently
- `pkg/` packages should be stable, well-documented, and potentially reusable

### `configs/` Directory

The `configs/` directory contains configuration files:

- **`config.yaml`**: The main configuration file for the application

This separation ensures that configuration is external to the code and can be modified without recompiling the application.

### `tests/` Directory

The `tests/` directory contains integration and end-to-end tests that may span multiple packages.

### `docs/` Directory

The `docs/` directory contains documentation files like this one.

## Design Principles

### Single Responsibility Principle

Each package has a single, well-defined responsibility:
- `camera` package: Everything related to IP cameras
- `network` package: Network scanning and RTSP verification
- `telegram` package: Everything related to Telegram integration
- `config` package: Configuration management
- `pkg/network` package: Generic network utilities

### Dependency Direction

The dependency direction follows clean architecture principles:
```
cmd/ -> internal/ -> pkg/
```

- `cmd/` depends on `internal/` packages
- `internal/` packages may depend on `pkg/` packages
- `pkg/` packages should not depend on `internal/` packages
- No circular dependencies are allowed

### Encapsulation

- `internal/` packages are not accessible from outside this module
- Only exported functions (starting with capital letter) are accessible
- Package-level variables and functions (starting with lowercase letter) are private to the package

## Benefits of This Structure

1. **Maintainability**: Clear separation makes it easy to locate and modify specific functionality
2. **Testability**: Well-defined interfaces make unit testing easier
3. **Reusability**: Generic functionality in `pkg/` can be reused in other projects
4. **Scalability**: New features can be added as new packages without disrupting existing code
5. **Collaboration**: Multiple developers can work on different packages simultaneously
6. **Security**: `internal/` packages cannot be accidentally imported by external projects

This structure follows Go community best practices and will make the project easier to maintain and extend as it grows.