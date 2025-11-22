# Application Architecture

This document describes the architecture of the Go IP Camera Scanner with Telegram Integration application.

## Overview

The application is designed to:
1. Discover IP cameras on a local network
2. Capture images from discovered cameras
3. Send captured images to Telegram

The architecture follows a modular design with clear separation of concerns.

## High-Level Architecture

```
┌─────────────────┐    ┌──────────────────┐    ┌──────────────────┐
│   Network Scan  │    │   RTSP Capture   │    │  Telegram Send   │
│                 │    │                 │    │                 │
│  Scan network   │───▶│  Connect to     │───▶│  Send image     │
│  for cameras    │    │  camera and    │    │  to Telegram     │
│                 │    │  capture image  │    │                 │
└─────────────────┘    └─────────────────┘    └──────────────────┘
         │                       │                       │
         └───────────────────────┼───────────────────────┘
                                 │
                      ┌──────────────────┐
                      │   Scheduler     │
                      │                 │
                      │  Control scan   │
                      │  intervals      │
                      └──────────────────┘
```

## Component Details

### 1. Configuration Manager (`internal/config`)

**Responsibilities**:
- Load configuration from YAML file
- Override with environment variables
- Validate configuration values
- Provide typed access to configuration

**Key Features**:
- Uses Viper for configuration management
- Supports multiple configuration sources
- Validates required fields
- Provides default values

### 2. Network Scanner (`internal/camera/scanner`)

**Responsibilities**:
- Parse CIDR notation to generate IP list
- Scan network for devices with open port 554
- Verify RTSP service availability
- Return list of discovered cameras

**Key Features**:
- Concurrent scanning with worker pools
- Configurable timeout settings
- Rate limiting to prevent network flooding
- Context-based cancellation support

### 3. RTSP Client (`internal/camera/rtsp`)

**Responsibilities**:
- Establish RTSP connections to cameras
- Handle authentication
- Manage RTSP session lifecycle
- Graceful connection handling

**Key Features**:
- Support for RTSP DESCRIBE, SETUP, PLAY commands
- Connection pooling for efficiency
- Proper resource cleanup
- Error handling and retry logic

### 4. Image Capture (`internal/camera/capture`)

**Responsibilities**:
- Capture single frames from RTSP streams
- Convert frames to JPEG format
- Handle different video codecs
- Manage memory efficiently

**Key Features**:
- Support for H.264/H.265 codecs
- Configurable image quality
- Timeout mechanisms
- Memory-efficient processing

### 5. Telegram Client (`internal/telegram`)

**Responsibilities**:
- Send images to Telegram
- Format messages with camera details
- Handle Telegram API responses
- Implement rate limiting

**Key Features**:
- Support for Telegram Bot API
- Retry logic for failed uploads
- Rate limiting compliance
- Error handling

### 6. Network Utilities (`pkg/network`)

**Responsibilities**:
- IP address manipulation
- Network connectivity checks
- CIDR parsing
- Generic network operations

**Key Features**:
- Reusable across projects
- Well-tested utility functions
- Efficient implementation

## Data Flow

1. **Initialization**:
   - Configuration manager loads settings
   - Dependencies are initialized

2. **Main Loop**:
   - Network scanner discovers cameras
   - For each discovered camera:
     - RTSP client connects to camera
     - Image capture extracts frame
     - Telegram client sends image
   - Wait for next scan interval

3. **Shutdown**:
   - Gracefully close all connections
   - Clean up resources
   - Exit application

## Concurrency Model

The application uses several concurrency patterns:

### Worker Pools
- Network scanning uses worker pools to limit concurrent connections
- Configurable number of workers based on system resources

### Goroutines
- Each camera processing runs in a separate goroutine
- Main loop runs in a goroutine for non-blocking operation

### Channels
- Used for communication between components
- Coordinate worker pool operations
- Handle cancellation signals

## Error Handling

The application implements comprehensive error handling:

### Retry Logic
- Network operations include retry mechanisms
- Exponential backoff with jitter
- Configurable retry limits

### Circuit Breaker
- Prevents repeated failures for persistently failing cameras
- Automatic recovery after timeout

### Graceful Degradation
- Failure of one camera doesn't affect others
- Continue operation with available cameras

## Resource Management

### Memory
- Efficient image processing to minimize memory usage
- Proper cleanup of temporary data
- Object pooling for frequently used structures

### Connections
- Connection pooling for RTSP connections
- Proper closing of network connections
- Resource limits to prevent exhaustion

### Files
- Temporary files are cleaned up after use
- Proper file handle management
- Secure handling of sensitive files

## Security Considerations

### Credential Management
- No hardcoded credentials
- Environment variable support
- Secure file permissions for sensitive data

### Input Validation
- Validate all external inputs
- Prevent injection attacks
- Validate network addresses and ranges

### Secure Communication
- RTSP authentication
- Secure Telegram API usage
- Proper error handling to prevent information leakage

## Performance Optimization

### Raspberry Pi Specific
- Limited concurrent operations to prevent resource exhaustion
- Efficient image processing
- Memory-conscious implementation

### Network Efficiency
- Rate limiting to prevent network flooding
- Connection pooling to reduce overhead
- Efficient scanning algorithms

### Resource Monitoring
- CPU and memory usage monitoring
- Automatic backoff when resources are constrained
- Logging for performance analysis

## Monitoring and Observability

### Logging
- Structured logging with levels
- Correlation IDs for request tracking
- Configurable log output

### Metrics
- Cameras discovered
- Successful captures
- Failed operations
- Performance metrics

### Health Checks
- Application health status
- Dependency availability
- Resource utilization

## Deployment Considerations

### Systemd Integration
- Automatic restart on failure
- Proper user permissions
- Environment variable loading
- Logging to journald

### Configuration
- External configuration files
- Environment variable overrides
- Secure credential handling

### Updates
- Binary replacement
- Configuration preservation
- Rollback procedures

## Future Extensibility

### Plugin Architecture
- Modular design allows for easy extension
- Well-defined interfaces for new features
- Minimal impact on existing code

### Configuration Expansion
- Easy addition of new configuration options
- Backward compatibility maintained
- Validation for new settings

### New Integrations
- Additional notification channels
- Different camera protocols
- Cloud storage options

## Conclusion

This architecture provides a solid foundation for the IP Camera Scanner application. The modular design with clear separation of concerns makes the application maintainable, testable, and extensible. The use of Go's concurrency features enables efficient processing of multiple cameras while the error handling and resource management ensure reliable operation.