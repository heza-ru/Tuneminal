# Tuneminal Windows Installation Script
# Requires PowerShell 5.0 or later

param(
    [string]$Version = "latest",
    [string]$InstallPath = "$env:USERPROFILE\AppData\Local\Programs\Tuneminal",
    [switch]$Force = $false
)

# Configuration
$REPO = "tuneminal/tuneminal"
$BINARY_NAME = "tuneminal"

# Functions
function Write-ColorOutput {
    param(
        [string]$Message,
        [string]$Color = "White"
    )
    Write-Host $Message -ForegroundColor $Color
}

function Write-Info {
    param([string]$Message)
    Write-ColorOutput "[INFO] $Message" "Cyan"
}

function Write-Success {
    param([string]$Message)
    Write-ColorOutput "[SUCCESS] $Message" "Green"
}

function Write-Warning {
    param([string]$Message)
    Write-ColorOutput "[WARNING] $Message" "Yellow"
}

function Write-Error {
    param([string]$Message)
    Write-ColorOutput "[ERROR] $Message" "Red"
}

function Get-LatestVersion {
    Write-Info "Fetching latest version..."
    
    try {
        $response = Invoke-RestMethod -Uri "https://api.github.com/repos/$REPO/releases/latest"
        $Version = $response.tag_name
        Write-Info "Latest version: $Version"
        return $Version
    }
    catch {
        Write-Error "Failed to fetch latest version: $($_.Exception.Message)"
        exit 1
    }
}

function Test-Dependencies {
    # Check if running as administrator
    $isAdmin = ([Security.Principal.WindowsPrincipal] [Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole] "Administrator")
    
    if (-not $isAdmin) {
        Write-Warning "Not running as administrator. Some operations may require elevated privileges."
    }
    
    # Check PowerShell version
    if ($PSVersionTable.PSVersion.Major -lt 5) {
        Write-Error "PowerShell 5.0 or later is required. Current version: $($PSVersionTable.PSVersion)"
        exit 1
    }
    
    Write-Info "Dependencies check passed"
}

function Install-Binary {
    param(
        [string]$Version,
        [string]$InstallPath
    )
    
    Write-Info "Installing Tuneminal to: $InstallPath"
    
    # Create install directory
    if (-not (Test-Path $InstallPath)) {
        New-Item -ItemType Directory -Path $InstallPath -Force | Out-Null
    }
    
    # Determine download URL
    $downloadUrl = "https://github.com/$REPO/releases/download/$Version/tuneminal-windows-amd64.exe"
    $binaryPath = Join-Path $InstallPath "tuneminal.exe"
    
    Write-Info "Downloading from: $downloadUrl"
    
    try {
        # Download binary
        Invoke-WebRequest -Uri $downloadUrl -OutFile $binaryPath -UseBasicParsing
        
        if (Test-Path $binaryPath) {
            Write-Success "Binary downloaded successfully"
        } else {
            throw "Downloaded file not found"
        }
    }
    catch {
        Write-Error "Failed to download binary: $($_.Exception.Message)"
        exit 1
    }
}

function Add-ToPath {
    param([string]$InstallPath)
    
    $binaryDir = Join-Path $InstallPath ""
    $currentPath = [Environment]::GetEnvironmentVariable("PATH", "User")
    
    if ($currentPath -notlike "*$binaryDir*") {
        Write-Info "Adding to PATH..."
        
        $newPath = if ($currentPath) { "$currentPath;$binaryDir" } else { $binaryDir }
        [Environment]::SetEnvironmentVariable("PATH", $newPath, "User")
        
        Write-Success "Added to PATH"
        Write-Warning "Please restart your terminal or refresh environment variables"
    } else {
        Write-Info "Already in PATH"
    }
}

function Install-DesktopShortcut {
    param([string]$InstallPath)
    
    $binaryPath = Join-Path $InstallPath "tuneminal.exe"
    $desktopPath = [Environment]::GetFolderPath("Desktop")
    $shortcutPath = Join-Path $desktopPath "Tuneminal.lnk"
    
    Write-Info "Creating desktop shortcut..."
    
    try {
        $WshShell = New-Object -comObject WScript.Shell
        $Shortcut = $WshShell.CreateShortcut($shortcutPath)
        $Shortcut.TargetPath = $binaryPath
        $Shortcut.WorkingDirectory = $InstallPath
        $Shortcut.Description = "Tuneminal - Command Line Karaoke Machine"
        $Shortcut.Save()
        
        Write-Success "Desktop shortcut created"
    }
    catch {
        Write-Warning "Failed to create desktop shortcut: $($_.Exception.Message)"
    }
}

function Install-StartMenuShortcut {
    param([string]$InstallPath)
    
    $binaryPath = Join-Path $InstallPath "tuneminal.exe"
    $startMenuPath = [Environment]::GetFolderPath("StartMenu")
    $programsPath = Join-Path $startMenuPath "Programs"
    $tuneminalPath = Join-Path $programsPath "Tuneminal"
    $shortcutPath = Join-Path $tuneminalPath "Tuneminal.lnk"
    
    Write-Info "Creating start menu shortcut..."
    
    try {
        if (-not (Test-Path $tuneminalPath)) {
            New-Item -ItemType Directory -Path $tuneminalPath -Force | Out-Null
        }
        
        $WshShell = New-Object -comObject WScript.Shell
        $Shortcut = $WshShell.CreateShortcut($shortcutPath)
        $Shortcut.TargetPath = $binaryPath
        $Shortcut.WorkingDirectory = $InstallPath
        $Shortcut.Description = "Tuneminal - Command Line Karaoke Machine"
        $Shortcut.Save()
        
        Write-Success "Start menu shortcut created"
    }
    catch {
        Write-Warning "Failed to create start menu shortcut: $($_.Exception.Message)"
    }
}

function Setup-Config {
    param([string]$InstallPath)
    
    $configPath = Join-Path $env:USERPROFILE ".tuneminal"
    
    Write-Info "Setting up configuration..."
    
    if (-not (Test-Path $configPath)) {
        New-Item -ItemType Directory -Path $configPath -Force | Out-Null
    }
    
    # Create default config
    $configFile = Join-Path $configPath "config.toml"
    if (-not (Test-Path $configFile)) {
        $configContent = @"
# Tuneminal Configuration

[audio]
# Default audio device (leave empty for system default)
device = ""

[visualizer]
# Number of bars in the visualizer
bars = 20
# Update frequency in milliseconds
update_freq = 100

[ui]
# Terminal theme
theme = "default"
# Show help text
show_help = true
"@
        Set-Content -Path $configFile -Value $configContent
        Write-Success "Created default configuration"
    }
    
    # Create demo directory
    $demoPath = Join-Path $configPath "demo"
    if (-not (Test-Path $demoPath)) {
        New-Item -ItemType Directory -Path $demoPath -Force | Out-Null
    }
    
    # Create demo lyrics file
    $demoLrcFile = Join-Path $demoPath "demo_song.lrc"
    if (-not (Test-Path $demoLrcFile)) {
        $demoContent = @"
[ar:Tuneminal Demo]
[ti:Welcome to Tuneminal]
[al:Demo Album]

[00:00.00]Welcome to Tuneminal
[00:03.50]Your command line karaoke machine
[00:07.00]Let's sing along together
[00:10.50]In this terminal scene

[00:14.00]Press space to play or pause
[00:17.50]Use Q to quit anytime
[00:21.00]Watch the visualizer dance
[00:24.50]To the rhythm and rhyme
"@
        Set-Content -Path $demoLrcFile -Value $demoContent
        Write-Success "Created demo lyrics file"
    }
    
    Write-Info "Configuration directory: $configPath"
    Write-Info "Demo files directory: $demoPath"
}

function Show-Usage {
    Write-ColorOutput @"
Tuneminal Windows Installation Script

Usage: .\install.ps1 [Options]

Options:
  -Version <version>     Version to install (default: latest)
  -InstallPath <path>    Installation directory (default: ~\AppData\Local\Programs\Tuneminal)
  -Force                 Force installation even if already installed
  -Help                  Show this help message

Examples:
  .\install.ps1                              # Install latest version
  .\install.ps1 -Version "v1.0.0"            # Install specific version
  .\install.ps1 -InstallPath "C:\Tuneminal"  # Custom installation path
  .\install.ps1 -Force                       # Force reinstallation

"@ "Cyan"
}

# Main installation process
function Main {
    if ($args -contains "-Help" -or $args -contains "--help" -or $args -contains "/?") {
        Show-Usage
        exit 0
    }
    
    Write-Info "Starting Tuneminal installation..."
    
    # Check dependencies
    Test-Dependencies
    
    # Get version
    if ($Version -eq "latest") {
        $Version = Get-LatestVersion
    }
    
    # Check if already installed
    $binaryPath = Join-Path $InstallPath "tuneminal.exe"
    if ((Test-Path $binaryPath) -and -not $Force) {
        Write-Warning "Tuneminal appears to be already installed at: $InstallPath"
        Write-Warning "Use -Force to reinstall"
        exit 0
    }
    
    # Install binary
    Install-Binary -Version $Version -InstallPath $InstallPath
    
    # Add to PATH
    Add-ToPath -InstallPath $InstallPath
    
    # Create shortcuts
    Install-DesktopShortcut -InstallPath $InstallPath
    Install-StartMenuShortcut -InstallPath $InstallPath
    
    # Setup configuration
    Setup-Config -InstallPath $InstallPath
    
    Write-Success "Installation completed successfully!"
    Write-Info "Run 'tuneminal' to start the application"
    Write-Info "Or use the desktop/start menu shortcuts"
}

# Run main function
Main $args





