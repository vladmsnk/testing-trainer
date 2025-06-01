#!/bin/bash

# HTTPS Setup Script for Testing Trainer
# This script helps you set up SSL certificates for your application

set -e

echo "🔒 Testing Trainer HTTPS Setup"
echo "=============================="

# Create certs directory if it doesn't exist
if [ ! -d "./certs" ]; then
    echo "📁 Creating certs directory..."
    mkdir -p ./certs
fi

# Check if certificates already exist
if [ -f "./certs/server.crt" ] && [ -f "./certs/server.key" ]; then
    echo "⚠️  SSL certificates already exist in ./certs/"
    read -p "Do you want to overwrite them? (y/N): " overwrite
    if [ "$overwrite" != "y" ] && [ "$overwrite" != "Y" ]; then
        echo "✅ Using existing certificates"
        exit 0
    fi
fi

echo "🔧 Generating SSL certificates..."

# Generate private key
echo "📝 Generating private key..."
openssl genrsa -out ./certs/server.key 2048

# Generate certificate
echo "📜 Generating self-signed certificate..."
openssl req -new -x509 -key ./certs/server.key -out ./certs/server.crt -days 365 \
    -subj "/C=US/ST=Development/L=Local/O=Testing Trainer/CN=localhost" \
    -addext "subjectAltName=DNS:localhost,IP:127.0.0.1,IP:::1,IP:0.0.0.0"

# Set appropriate permissions
chmod 600 ./certs/server.key
chmod 644 ./certs/server.crt

echo "✅ SSL certificates generated successfully!"
echo ""
echo "📋 Certificate Information:"
echo "   Certificate: ./certs/server.crt"
echo "   Private Key: ./certs/server.key"
echo "   Valid for: 365 days"
echo "   Domains: localhost, 127.0.0.1"
echo ""

# Check if environment variables are set
echo "🔍 Checking environment configuration..."
if [ -z "$TLS_CERT_FILE" ] && [ -z "$TLS_KEY_FILE" ]; then
    echo "💡 Setting environment variables for this session:"
    export TLS_CERT_FILE="./certs/server.crt"
    export TLS_KEY_FILE="./certs/server.key"
    echo "   TLS_CERT_FILE=./certs/server.crt"
    echo "   TLS_KEY_FILE=./certs/server.key"
    echo ""
    echo "⚠️  To make these permanent, add them to your shell profile:"
    echo "   echo 'export TLS_CERT_FILE=\"./certs/server.crt\"' >> ~/.bashrc"
    echo "   echo 'export TLS_KEY_FILE=\"./certs/server.key\"' >> ~/.bashrc"
    echo "   source ~/.bashrc"
else
    echo "✅ Environment variables already set:"
    echo "   TLS_CERT_FILE=$TLS_CERT_FILE"
    echo "   TLS_KEY_FILE=$TLS_KEY_FILE"
fi

echo ""
echo "🚀 Ready to start your application!"
echo ""
echo "📖 Next steps:"
echo "1. Start your application: go run cmd/app/main.go"
echo "2. HTTP will be available at: http://localhost:7001"
echo "3. HTTPS will be available at: https://localhost:8001"
echo "4. Swagger UI: https://localhost:8001/swagger/index.html"
echo ""
echo "⚠️  Browser Security Warning:"
echo "   Your browser will show a security warning for self-signed certificates."
echo "   Click 'Advanced' → 'Proceed to localhost (unsafe)' to continue."
echo ""
echo "📚 For production setup, see: SSL_SETUP_GUIDE.md" 