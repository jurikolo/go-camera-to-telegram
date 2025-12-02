# Building the Application

This document explains how to build and run the IP Camera Scanner application using the provided Makefile.

## Prerequisites

- Go 1.25 or later
- Make (optional, but recommended)

## Using the Makefile

The project includes a Makefile with several useful targets:

### Build for Current Platform

```bash
make build
```

This builds the application for your current platform and creates a `scanner` binary.

### Build for Raspberry Pi

```bash
make build-pi
```

This cross-compiles the application for Raspberry Pi (ARM64) and creates a `scanner-arm64` binary.

### Install Dependencies

```bash
make deps
```

This downloads and installs all required Go dependencies.

### Run Tests

```bash
make test
```

This runs all unit and integration tests.

### Clean Build Artifacts

```bash
make clean
```

This removes all build artifacts and binaries.

### Run the Application

```bash
make run
```

This builds and runs the application.

### Install the Binary

```bash
make install
```

This builds the application and installs the binary to `/usr/local/bin`.

### Show Help

```bash
make help
```

This displays information about all available Makefile targets.

## Manual Building

If you prefer to build manually without Make:

### Build for Current Platform

```bash
go build -tags 'netgo osusergo' -a -installsuffix cgo -o scanner ./cmd/scanner
```

### Build for Raspberry Pi

```bash
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -a -installsuffix cgo -o scanner-arm64 ./cmd/scanner
```

### Run Tests

```bash
go test ./...
```

## Running the Application

After building, you can run the application directly:

```bash
./scanner
```

To see available command-line options:

```bash
./scanner --help
```

## Configuration

The application can be configured through:

1. The `configs/config.yaml` file
2. Environment variables
3. Command-line arguments (when implemented)

## Systemd Service

For running the application as a service on Linux systems, you can use the provided systemd service file in `deploy/scanner.service`.

To install as a service:

1. Copy the binary to an appropriate location:
   ```bash
   sudo cp scanner /usr/local/bin/
   ```

2. Copy the service file:
   ```bash
   sudo cp deploy/scanner.service /etc/systemd/system/
   ```

3. Edit the service file to match your configuration:
   ```bash
   sudo nano /etc/systemd/system/scanner.service
   ```

4. Enable and start the service:
   ```bash
   sudo systemctl enable scanner.service
   sudo systemctl start scanner.service
   ```

## Cross-Platform Building

The Makefile supports cross-platform building. To build for other platforms, you can modify the Makefile or use Go's built-in cross-compilation:

### Build for Windows

```bash
GOOS=windows GOARCH=amd64 go build -o scanner.exe ./cmd/scanner
```

### Build for macOS

```bash
GOOS=darwin GOARCH=amd64 go build -o scanner-mac ./cmd/scanner
```

## Troubleshooting

### Missing Dependencies

If you encounter dependency issues:

```bash
make deps
```

or

```bash
go mod tidy
```

### Permission Issues

If you encounter permission issues when installing:

```bash
sudo make install
```
