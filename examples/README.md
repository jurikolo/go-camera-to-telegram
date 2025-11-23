# Examples

This directory contains example code demonstrating how to use the various components of the IP Camera Scanner application.

## Camera Examples

### Basic Camera Scanning

```bash
cd examples/camera
go run ../examples/camera/camera_scan.go
```

### RTSP Client Usage

```bash
cd examples/camera
go run rtsp_client.go <host> <username> <password>
```

Example:
```bash
go run rtsp_client.go 192.168.1.100 admin password
```

## Network Examples

### RTSP Verification

```bash
cd examples/network
go run rtsp_verification.go <host>
```

Example:
```bash
go run rtsp_verification.go 127.0.0.1
```

## Note

These examples are for demonstration purposes only and may require actual network access and IP cameras to function properly. Modify the configuration and code as needed for your specific environment.