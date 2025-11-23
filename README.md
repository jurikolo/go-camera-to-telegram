# IP Camera Scanner with Telegram Integration

This project scans a network for IP cameras, captures images from them via RTSP, and sends the images to a Telegram chat.

## Features

- Network scanning for IP cameras on port 554 (RTSP)
- RTSP connection and frame capture with H.264/H.265 codec support
- JPEG image conversion with quality and resolution options
- Telegram integration for sending captured images
- Configuration via YAML file or environment variables
- Connection pooling for efficient resource management

## Getting Started

### Prerequisites

- Go 1.21 or later
- Network access to IP cameras with RTSP support
- Telegram bot token and chat ID

### Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/yourusername/go-camera-to-telegram.git
   cd go-camera-to-telegram
   ```

2. Install dependencies:
   ```bash
   go mod tidy
   ```

3. Configure the application (see Configuration section below)

4. Build the application:
   ```bash
   go build -o scanner ./cmd/scanner
   ```

### Configuration

Create a `config.yaml` file in the `configs` directory or in the root directory:

```yaml
# Network configuration
network:
  # CIDR notation for network scanning
  cidr: "192.168.1.0/24"
  # Timeout in seconds for network operations
  timeout: 5
  # Number of concurrent workers for scanning
  workers: 10

# RTSP configuration
rtsp:
  # Default username for RTSP authentication
  username: "admin"
  # Default password for RTSP authentication
  password: "password"
  # Timeout in seconds for RTSP operations
  timeout: 30

# Telegram configuration
telegram:
  # Telegram bot token (get from @BotFather)
  bot_token: "YOUR_BOT_TOKEN_HERE"
  # Chat ID to send messages to
  chat_id: "YOUR_CHAT_ID_HERE"

# Scan configuration
scan:
  # Interval between scans in minutes
  interval: 30
  # Maximum number of cameras to process concurrently
  max_concurrent: 3
```

You can also configure the application using environment variables:
- `CTT_NETWORK_CIDR`
- `CTT_NETWORK_TIMEOUT`
- `CTT_NETWORK_WORKERS`
- `CTT_RTSP_USERNAME`
- `CTT_RTSP_PASSWORD`
- `CTT_RTSP_TIMEOUT`
- `CTT_TELEGRAM_BOT_TOKEN`
- `CTT_TELEGRAM_CHAT_ID`
- `CTT_SCAN_INTERVAL`
- `CTT_SCAN_MAX_CONCURRENT`

### Usage

Run the scanner:
```bash
./scanner
```

Or with a specific config file:
```bash
./scanner --config /path/to/config.yaml
```

### Image Capture

The application can capture single frames from RTSP streams and convert them to JPEG format. The capture functionality supports:

- H.264 and H.265 codec handling
- Configurable timeout mechanisms
- JPEG quality settings (1-100)
- Maximum width and height constraints
- Connection pooling for efficient resource management

Example usage:
```go
// Create camera capture instance
capture := camera.NewCapture()

// Set capture options
options := &camera.CaptureOptions{
    Timeout:   30 * time.Second,
    Quality:   75,
    MaxWidth:  1920,
    MaxHeight: 1080,
}

// Capture frame
jpegData, err := capture.CaptureFrame(cameraIP, username, password, options)
```

## Project Structure

```
.
├── cmd/
│   └── scanner/
│       └── main.go          # Main application entry point
├── configs/
│   └── config.yaml           # Configuration file
├── deploy/
│   └── scanner.service      # Systemd service file
├── docs/
│   ├── ARCHITECTURE.md      # Architecture documentation
│   ├── BUILDING.md          # Build instructions
│   ├── IMPLEMENTATION_SUMMARY.md
│   ├── PROJECT_OVERVIEW.md
│   ├── PROJECT_STRUCTURE.md
│   └── SEPARATION_OF_CONCERNS.md
├── examples/
│   └── capture_example.go    # Example of image capture usage
├── internal/
│   ├── camera/
│   │   ├── capture.go       # Image capture functionality
│   │   ├── pool.go          # Connection pool for RTSP clients
│   │   ├── rtsp.go          # RTSP client implementation
│   │   ├── rtsp_test.go     # RTSP client tests
│   │   └── scanner.go       # Network scanner for IP cameras
│   ├── config/
│   │   └── config.go       # Configuration management
│   ├── network/
│   │   ├── scanner.go        # Network scanning utilities
│   │   └── scanner_test.go   # Network scanner tests
│   └── telegram/
│       └── client.go         # Telegram client
├── tests/
│   └── integration_test.go   # Integration tests
├── go.mod
├── go.sum
├── LICENSE
├── Makefile
└── README.md
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.