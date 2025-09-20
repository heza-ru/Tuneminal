#!/bin/bash

# Tuneminal Installation Script
# Supports Linux, macOS, and Windows (via WSL/Git Bash)

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
REPO="tuneminal/tuneminal"
BINARY_NAME="tuneminal"
INSTALL_DIR="${HOME}/.local/bin"
CONFIG_DIR="${HOME}/.config/tuneminal"

# Functions
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

detect_os() {
    if [[ "$OSTYPE" == "linux-gnu"* ]]; then
        OS="linux"
    elif [[ "$OSTYPE" == "darwin"* ]]; then
        OS="darwin"
    elif [[ "$OSTYPE" == "cygwin" ]] || [[ "$OSTYPE" == "msys" ]]; then
        OS="windows"
    else
        print_error "Unsupported operating system: $OSTYPE"
        exit 1
    fi
    print_info "Detected OS: $OS"
}

detect_arch() {
    ARCH=$(uname -m)
    case $ARCH in
        x86_64)
            ARCH="amd64"
            ;;
        arm64|aarch64)
            ARCH="arm64"
            ;;
        armv7l)
            ARCH="armv7"
            ;;
        *)
            print_error "Unsupported architecture: $ARCH"
            exit 1
            ;;
    esac
    print_info "Detected architecture: $ARCH"
}

get_latest_version() {
    print_info "Fetching latest version..."
    VERSION=$(curl -s "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')
    if [ -z "$VERSION" ]; then
        print_error "Failed to fetch latest version"
        exit 1
    fi
    print_info "Latest version: $VERSION"
}

download_binary() {
    local url="https://github.com/$REPO/releases/download/$VERSION/$BINARY_NAME-$OS-$ARCH"
    if [ "$OS" = "windows" ]; then
        url="${url}.exe"
    fi
    
    print_info "Downloading from: $url"
    
    # Create temp directory
    TEMP_DIR=$(mktemp -d)
    cd "$TEMP_DIR"
    
    # Download binary
    if command -v wget >/dev/null 2>&1; then
        wget -q "$url" -O "$BINARY_NAME"
    elif command -v curl >/dev/null 2>&1; then
        curl -sL "$url" -o "$BINARY_NAME"
    else
        print_error "Neither wget nor curl found. Please install one of them."
        exit 1
    fi
    
    if [ ! -f "$BINARY_NAME" ]; then
        print_error "Failed to download binary"
        exit 1
    fi
    
    # Make executable
    chmod +x "$BINARY_NAME"
    
    print_success "Binary downloaded successfully"
}

install_binary() {
    # Create install directory
    mkdir -p "$INSTALL_DIR"
    
    # Copy binary
    cp "$BINARY_NAME" "$INSTALL_DIR/"
    
    # Add to PATH if not already present
    if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
        print_warning "Adding $INSTALL_DIR to PATH"
        
        # Detect shell
        if [ -n "$ZSH_VERSION" ]; then
            SHELL_RC="$HOME/.zshrc"
        elif [ -n "$BASH_VERSION" ]; then
            SHELL_RC="$HOME/.bashrc"
        else
            SHELL_RC="$HOME/.profile"
        fi
        
        echo "export PATH=\"$INSTALL_DIR:\$PATH\"" >> "$SHELL_RC"
        print_info "Added to PATH in $SHELL_RC"
        print_warning "Please restart your terminal or run: source $SHELL_RC"
    fi
    
    print_success "Binary installed to $INSTALL_DIR"
}

create_config() {
    # Create config directory
    mkdir -p "$CONFIG_DIR"
    
    # Create default config if it doesn't exist
    if [ ! -f "$CONFIG_DIR/config.toml" ]; then
        cat > "$CONFIG_DIR/config.toml" << EOF
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
EOF
        print_success "Created default configuration at $CONFIG_DIR/config.toml"
    fi
}

setup_demo_files() {
    # Create demo directory
    DEMO_DIR="$CONFIG_DIR/demo"
    mkdir -p "$DEMO_DIR"
    
    # Copy demo files if they don't exist
    if [ ! -f "$DEMO_DIR/demo_song.lrc" ]; then
        cat > "$DEMO_DIR/demo_song.lrc" << 'EOF'
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
EOF
        print_success "Created demo lyrics file"
    fi
    
    print_info "Demo files available at: $DEMO_DIR"
    print_warning "Add your own MP3/WAV files to this directory to get started!"
}

cleanup() {
    if [ -n "$TEMP_DIR" ] && [ -d "$TEMP_DIR" ]; then
        rm -rf "$TEMP_DIR"
    fi
}

# Main installation process
main() {
    print_info "Starting Tuneminal installation..."
    
    # Detect system
    detect_os
    detect_arch
    
    # Get latest version
    get_latest_version
    
    # Download and install
    download_binary
    install_binary
    
    # Setup configuration
    create_config
    setup_demo_files
    
    # Cleanup
    cleanup
    
    print_success "Installation completed successfully!"
    print_info "Run 'tuneminal' to start the application"
    print_info "Configuration directory: $CONFIG_DIR"
    print_info "Demo files directory: $CONFIG_DIR/demo"
}

# Handle script interruption
trap cleanup EXIT

# Run main function
main "$@"

