# MamaCare SL - Hasura Development Environment Startup Script
# This script initializes the local development environment for Hasura

# Ensure Docker is running
Write-Host "Checking Docker status..." -ForegroundColor Cyan
$dockerStatus = docker info 2>&1
if ($LASTEXITCODE -ne 0) {
    Write-Host "Error: Docker is not running. Please start Docker Desktop first." -ForegroundColor Red
    exit 1
}
Write-Host "Docker is running!" -ForegroundColor Green

# Check if .env file exists
if (-not (Test-Path -Path ".env")) {
    Write-Host "Warning: .env file not found. Creating from template..." -ForegroundColor Yellow
    Copy-Item -Path "env.template" -Destination ".env"
    Write-Host ".env file created. Please update with proper secrets before deploying to production." -ForegroundColor Yellow
}

# Start hasura and postgres containers
Write-Host "Starting Hasura and PostgreSQL containers..." -ForegroundColor Cyan
docker-compose down
docker-compose up -d

# Wait for services to be ready
Write-Host "Waiting for services to be fully operational..." -ForegroundColor Cyan
Start-Sleep -Seconds 10

# Check if hasura CLI is installed
$hasuraCLI = Get-Command hasura -ErrorAction SilentlyContinue
if (-not $hasuraCLI) {
    Write-Host "Hasura CLI not detected. Please install it with: npm install --global hasura-cli" -ForegroundColor Yellow
    Write-Host "Then run: hasura console --project services/hasura --admin-secret <your-admin-secret>" -ForegroundColor Yellow
} else {
    # Launch Hasura console
    Write-Host "Launching Hasura console..." -ForegroundColor Green
    
    # Get admin secret from .env file
    $adminSecret = "mamacare-dev-admin-secret"
    if (Test-Path -Path ".env") {
        $envContent = Get-Content ".env"
        foreach ($line in $envContent) {
            if ($line -match "HASURA_GRAPHQL_ADMIN_SECRET=(.*)") {
                $adminSecret = $matches[1]
                break
            }
        }
    }
    
    Write-Host "Open your browser to http://localhost:8080 and use the admin secret to log in" -ForegroundColor Green
    Write-Host "To launch the console with CLI features, run: hasura console --admin-secret $adminSecret" -ForegroundColor Cyan
}

Write-Host "Development environment is ready!" -ForegroundColor Green
