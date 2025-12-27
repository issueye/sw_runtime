# PowerShell script to generate self-signed SSL certificate
# Usage: .\generate-cert.ps1

Write-Host "Generating self-signed SSL certificate for HTTPS demo..." -ForegroundColor Green
Write-Host ""

# Check if OpenSSL is available
$opensslPath = Get-Command openssl -ErrorAction SilentlyContinue

if (-not $opensslPath) {
    Write-Host "ERROR: OpenSSL is not installed or not in PATH" -ForegroundColor Red
    Write-Host ""
    Write-Host "Please install OpenSSL:" -ForegroundColor Yellow
    Write-Host "  Download from: https://slproweb.com/products/Win32OpenSSL.html" -ForegroundColor Yellow
    Write-Host "  Or use Chocolatey: choco install openssl" -ForegroundColor Yellow
    Write-Host ""
    exit 1
}

Write-Host "OpenSSL found: $($opensslPath.Source)" -ForegroundColor Cyan
Write-Host ""

# Certificate details
$certSubject = "/C=CN/ST=Beijing/L=Beijing/O=SW-Runtime/OU=Development/CN=localhost"
$validDays = 365

# Generate private key
Write-Host "Step 1: Generating private key..." -ForegroundColor Yellow
openssl genrsa -out server.key 2048
if ($LASTEXITCODE -ne 0) {
    Write-Host "Failed to generate private key" -ForegroundColor Red
    exit 1
}
Write-Host "✓ Private key generated: server.key" -ForegroundColor Green
Write-Host ""

# Generate certificate
Write-Host "Step 2: Generating self-signed certificate..." -ForegroundColor Yellow
openssl req -x509 -new -key server.key -out server.crt -days $validDays -subj $certSubject
if ($LASTEXITCODE -ne 0) {
    Write-Host "Failed to generate certificate" -ForegroundColor Red
    exit 1
}
Write-Host "✓ Certificate generated: server.crt (valid for $validDays days)" -ForegroundColor Green
Write-Host ""

# Display certificate info
Write-Host "Certificate Information:" -ForegroundColor Cyan
Write-Host "========================" -ForegroundColor Cyan
openssl x509 -in server.crt -text -noout | Select-String "Subject:", "Not Before", "Not After"
Write-Host ""

Write-Host "✓ SSL Certificate generation complete!" -ForegroundColor Green
Write-Host ""
Write-Host "Files generated:" -ForegroundColor Cyan
Write-Host "  - server.key (Private Key)" -ForegroundColor White
Write-Host "  - server.crt (Certificate)" -ForegroundColor White
Write-Host ""
Write-Host "Usage in your code:" -ForegroundColor Yellow
Write-Host "  app.listenTLS('8443', './examples/certs/server.crt', './examples/certs/server.key')" -ForegroundColor White
Write-Host ""
Write-Host "⚠️  Note: This is a self-signed certificate for development only." -ForegroundColor Yellow
Write-Host "    Browsers will show security warnings. This is expected." -ForegroundColor Yellow
Write-Host ""
