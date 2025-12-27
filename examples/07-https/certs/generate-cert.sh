#!/bin/bash
# Shell script to generate self-signed SSL certificate
# Usage: ./generate-cert.sh

echo "Generating self-signed SSL certificate for HTTPS demo..."
echo ""

# Check if OpenSSL is available
if ! command -v openssl &> /dev/null; then
    echo "ERROR: OpenSSL is not installed"
    echo ""
    echo "Please install OpenSSL:"
    echo "  Ubuntu/Debian: sudo apt-get install openssl"
    echo "  CentOS/RHEL:   sudo yum install openssl"
    echo "  macOS:         brew install openssl"
    echo ""
    exit 1
fi

echo "OpenSSL found: $(which openssl)"
echo ""

# Certificate details
CERT_SUBJECT="/C=CN/ST=Beijing/L=Beijing/O=SW-Runtime/OU=Development/CN=localhost"
VALID_DAYS=365

# Generate private key and certificate
echo "Generating private key and certificate..."
openssl req -x509 -newkey rsa:2048 -nodes -keyout server.key -out server.crt -days $VALID_DAYS \
  -subj "$CERT_SUBJECT"

if [ $? -ne 0 ]; then
    echo "Failed to generate certificate"
    exit 1
fi

echo ""
echo "✓ SSL Certificate generation complete!"
echo ""

# Display certificate info
echo "Certificate Information:"
echo "========================"
openssl x509 -in server.crt -text -noout | grep -E "Subject:|Not Before|Not After"
echo ""

echo "Files generated:"
echo "  - server.key (Private Key)"
echo "  - server.crt (Certificate)"
echo ""
echo "Usage in your code:"
echo "  app.listenTLS('8443', './examples/certs/server.crt', './examples/certs/server.key')"
echo ""
echo "⚠️  Note: This is a self-signed certificate for development only."
echo "    Browsers will show security warnings. This is expected."
echo ""

# Make the script executable
chmod +x generate-cert.sh
