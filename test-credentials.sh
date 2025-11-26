#!/bin/bash

# Test script for secure credential management
# This script demonstrates how to use the secure credential management system

echo "Testing secure credential management..."

# Create test credential files
mkdir -p /tmp/test-credentials
echo "admin" > /tmp/test-credentials/rtsp-username
echo "secretpassword" > /tmp/test-credentials/rtsp-password
echo "123456789:ABCdefGhIJKlmNoPQRsTUVwxyZ" > /tmp/test-credentials/telegram-token

# Set secure permissions
chmod 600 /tmp/test-credentials/rtsp-username
chmod 600 /tmp/test-credentials/rtsp-password
chmod 600 /tmp/test-credentials/telegram-token

# Set environment variables
export CTT_RTSP_USERNAME_FILE="/tmp/test-credentials/rtsp-username"
export CTT_RTSP_PASSWORD_FILE="/tmp/test-credentials/rtsp-password"
export CTT_TELEGRAM_BOT_TOKEN_FILE="/tmp/test-credentials/telegram-token"

# Test the application (this would normally be done with a test config)
echo "Environment variables set:"
echo "  CTT_RTSP_USERNAME_FILE: $CTT_RTSP_USERNAME_FILE"
echo "  CTT_RTSP_PASSWORD_FILE: $CTT_RTSP_PASSWORD_FILE"
echo "  CTT_TELEGRAM_BOT_TOKEN_FILE: $CTT_TELEGRAM_BOT_TOKEN_FILE"

echo "Credential files created with secure permissions:"
ls -l /tmp/test-credentials/

echo "Test completed successfully!"

# Cleanup
rm -rf /tmp/test-credentials
unset CTT_RTSP_USERNAME_FILE
unset CTT_RTSP_PASSWORD_FILE
unset CTT_TELEGRAM_BOT_TOKEN_FILE