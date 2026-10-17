#!/usr/bin/env bash
#
# Coffee Bun (bun) installer for macOS and Linux
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/KennethanCeyer/bundle/main/install.sh | bash
#

set -e

# ANSI color codes
BOLD="$(tput bold 2>/dev/null || printf '')"
GREEN="$(tput setaf 2 2>/dev/null || printf '')"
BLUE="$(tput setaf 4 2>/dev/null || printf '')"
YELLOW="$(tput setaf 3 2>/dev/null || printf '')"
RED="$(tput setaf 1 2>/dev/null || printf '')"
RESET="$(tput sgr0 2>/dev/null || printf '')"

printf "\n%sCoffee Bun (bun) - Cross-Platform Bootstrap Installer%s\n\n" "${BOLD}${BLUE}" "${RESET}"

# 1. Check OS
OS="$(uname -s)"
ARCH="$(uname -m)"
if [ "$OS" != "Darwin" ] && [ "$OS" != "Linux" ]; then
    printf "%s[!] Unsupported Unix OS: %s (For Windows, run install.ps1 in PowerShell)%s\n" "${RED}" "$OS" "${RESET}"
    exit 1
fi

printf "  %s[OK]%s Operating System: %s (%s)\n" "${GREEN}" "${RESET}" "$OS" "$ARCH"

# 2. Check Package Manager / Go
if ! command -v go >/dev/null 2>&1; then
    printf "  %s[..]%s Go not detected. Attempting to install Go...\n" "${YELLOW}" "${RESET}"
    if [ "$OS" = "Darwin" ]; then
        if ! command -v brew >/dev/null 2>&1; then
            /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
            if [ -f "/opt/homebrew/bin/brew" ]; then
                eval "$(/opt/homebrew/bin/brew shellenv)"
            elif [ -f "/usr/local/bin/brew" ]; then
                eval "$(/usr/local/bin/brew shellenv)"
            fi
        fi
        brew install go
    elif command -v apt-get >/dev/null 2>&1; then
        sudo apt-get update -y && sudo apt-get install -y golang-go git
    elif command -v dnf >/dev/null 2>&1; then
        sudo dnf install -y golang git
    elif command -v apk >/dev/null 2>&1; then
        sudo apk add --no-cache go git
    else
        printf "%s[!] Please install Go (https://go.dev/dl/) and re-run this installer.%s\n" "${RED}" "${RESET}"
        exit 1
    fi
else
    printf "  %s[OK]%s Go is ready (%s)\n" "${GREEN}" "${RESET}" "$(go version | awk '{print $3}')"
fi

# 3. Clone or build Coffee Bun
INSTALL_DIR="$HOME/.coffee-bun"
BIN_DIR="$HOME/.local/bin"
mkdir -p "$BIN_DIR"

if [ -d "$PWD/.git" ] && [ -f "$PWD/go.mod" ]; then
    SRC_DIR="$PWD"
    printf "  %s[..]%s Building Coffee Bun from current repository...\n" "${BLUE}" "${RESET}"
else
    SRC_DIR="$INSTALL_DIR"
    if [ -d "$SRC_DIR" ]; then
        printf "  %s[..]%s Updating existing Coffee Bun source at %s...\n" "${BLUE}" "${RESET}" "$SRC_DIR"
        git -C "$SRC_DIR" pull --rebase
    else
        printf "  %s[..]%s Cloning Coffee Bun repository to %s...\n" "${BLUE}" "${RESET}" "$SRC_DIR"
        git clone https://github.com/KennethanCeyer/bundle.git "$SRC_DIR"
    fi
fi

# 4. Build binary
rm -f "$BIN_DIR/bun" "$BIN_DIR/coffee-bun"
(cd "$SRC_DIR" && go build -ldflags="-s -w" -o "$BIN_DIR/bun" ./cmd/bun)
if [ "$OS" = "Darwin" ] && command -v codesign >/dev/null 2>&1; then
    codesign -s - -f "$BIN_DIR/bun" 2>/dev/null || true
fi
ln -sf "$BIN_DIR/bun" "$BIN_DIR/coffee-bun"

# 5. Ensure ~/.local/bin is in PATH in shell rc files
for RC_FILE in "$HOME/.zshrc" "$HOME/.zprofile" "$HOME/.bashrc" "$HOME/.profile"; do
    if [ -f "$RC_FILE" ]; then
        if ! grep -q 'HOME/.local/bin' "$RC_FILE" 2>/dev/null; then
            printf "\n# Added by Coffee Bun\nexport PATH=\"\$HOME/.local/bin:\$PATH\"\n" >> "$RC_FILE"
        fi
    fi
done

printf "\n%s[OK] Coffee Bun installed successfully to %s/bun!%s\n\n" "${GREEN}${BOLD}" "$BIN_DIR" "${RESET}"
printf "To launch interactive setup, simply run:\n"
printf "  %s%sbun%s (or %scoffee-bun%s)\n\n" "${BOLD}" "${BLUE}" "${RESET}" "${BLUE}" "${RESET}"
printf "Or inspect system status:\n"
printf "  %s%sbun check%s\n\n" "${BOLD}" "${BLUE}" "${RESET}"
