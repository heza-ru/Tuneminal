# Tuneminal PowerShell Build Script

param(
    [string]$Target = "build",
    [switch]$Release = $false,
    [switch]$Clean = $false
)

function Build-Application {
    Write-Host "Building Tuneminal..." -ForegroundColor Green
    
    if ($Release) {
        Write-Host "Building optimized release version..." -ForegroundColor Yellow
        go build -ldflags="-s -w" -o tuneminal.exe cmd/tuneminal/main.go
    } else {
        Write-Host "Building development version..." -ForegroundColor Yellow
        go build -o tuneminal.exe cmd/tuneminal/main.go
    }
    
    if ($LASTEXITCODE -eq 0) {
        Write-Host "✅ Build successful!" -ForegroundColor Green
        Write-Host "Binary created: tuneminal.exe" -ForegroundColor Cyan
    } else {
        Write-Host "❌ Build failed!" -ForegroundColor Red
        exit 1
    }
}

function Run-Application {
    Write-Host "Running Tuneminal..." -ForegroundColor Green
    
    if (Test-Path "tuneminal.exe") {
        .\tuneminal.exe
    } else {
        Write-Host "❌ tuneminal.exe not found. Run build first." -ForegroundColor Red
        exit 1
    }
}

function Test-Application {
    Write-Host "Running tests..." -ForegroundColor Green
    go test ./...
    
    if ($LASTEXITCODE -eq 0) {
        Write-Host "✅ All tests passed!" -ForegroundColor Green
    } else {
        Write-Host "❌ Some tests failed!" -ForegroundColor Red
        exit 1
    }
}

function Install-Dependencies {
    Write-Host "Installing dependencies..." -ForegroundColor Green
    go mod tidy
    go mod download
    
    if ($LASTEXITCODE -eq 0) {
        Write-Host "✅ Dependencies installed!" -ForegroundColor Green
    } else {
        Write-Host "❌ Failed to install dependencies!" -ForegroundColor Red
        exit 1
    }
}

function Clean-Artifacts {
    Write-Host "Cleaning build artifacts..." -ForegroundColor Green
    
    if (Test-Path "tuneminal.exe") {
        Remove-Item "tuneminal.exe" -Force
        Write-Host "Removed tuneminal.exe" -ForegroundColor Yellow
    }
    
    go clean
    
    Write-Host "✅ Cleanup complete!" -ForegroundColor Green
}

function Show-Help {
    Write-Host @"
Tuneminal Build Script

Usage: .\build.ps1 [Target] [Options]

Targets:
  build      Build the application (default)
  run        Build and run the application
  test       Run tests
  deps       Install dependencies
  clean      Clean build artifacts
  release    Build optimized release version
  help       Show this help

Options:
  -Release   Build optimized release version
  -Clean     Clean before building

Examples:
  .\build.ps1                    # Build development version
  .\build.ps1 run               # Build and run
  .\build.ps1 release -Release  # Build optimized release
  .\build.ps1 clean             # Clean artifacts
  .\build.ps1 test              # Run tests

"@ -ForegroundColor Cyan
}

# Main execution
switch ($Target.ToLower()) {
    "build" {
        if ($Clean) { Clean-Artifacts }
        Build-Application
    }
    "run" {
        if ($Clean) { Clean-Artifacts }
        Build-Application
        Run-Application
    }
    "test" {
        Test-Application
    }
    "deps" {
        Install-Dependencies
    }
    "clean" {
        Clean-Artifacts
    }
    "release" {
        if ($Clean) { Clean-Artifacts }
        Build-Application -Release
    }
    "help" {
        Show-Help
    }
    default {
        Write-Host "Unknown target: $Target" -ForegroundColor Red
        Show-Help
        exit 1
    }
}

