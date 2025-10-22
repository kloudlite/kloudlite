#!/bin/bash
# Initialize shell configuration files for kl user
# This script runs on container startup to ensure shell configs are properly set up
# even when home directory is persisted from a volume

set -e

KL_HOME="/home/kl"
BASHRC="$KL_HOME/.bashrc"
BASH_PROFILE="$KL_HOME/.bash_profile"
ZSHRC="$KL_HOME/.zshrc"
FISH_CONFIG="$KL_HOME/.config/fish/config.fish"
STARSHIP_CONFIG="$KL_HOME/.config/starship.toml"

# Marker to identify our configuration
BASHRC_MARKER="# Kloudlite shell initialization"
BASH_PROFILE_MARKER="# Source .bashrc for login shells"
TERM_FIX_MARKER="# Fix TERM for SSH login"

echo "Initializing shell configurations for kl user..."

# Function to check if a marker exists in a file
has_marker() {
    local file="$1"
    local marker="$2"
    [ -f "$file" ] && grep -q "$marker" "$file"
}

# Initialize bash configuration
if ! has_marker "$BASHRC" "$BASHRC_MARKER"; then
    echo "Setting up .bashrc..."
    cat >> "$BASHRC" << 'EOF'
# Fix TERM for SSH login
case "$TERM" in
  xterm-kitty|xterm-256color|screen-256color)
    export TERM=xterm-256color
    ;;
  xterm*)
    export TERM=xterm
    ;;
esac

# Kloudlite shell initialization
source /etc/kloudlite/shell-init.sh
EOF
    chown kl:kl "$BASHRC"
    echo ".bashrc configured"
else
    echo ".bashrc already configured"
fi

# Initialize bash_profile to source bashrc for SSH login shells
if ! has_marker "$BASH_PROFILE" "$BASH_PROFILE_MARKER"; then
    echo "Setting up .bash_profile..."
    cat >> "$BASH_PROFILE" << 'EOF'
# Source .bashrc for login shells
if [ -f ~/.bashrc ]; then
    source ~/.bashrc
fi
EOF
    chown kl:kl "$BASH_PROFILE"
    echo ".bash_profile configured"
else
    echo ".bash_profile already configured"
fi

# Initialize zsh configuration
if ! has_marker "$ZSHRC" "$BASHRC_MARKER"; then
    echo "Setting up .zshrc..."
    cat > "$ZSHRC" << 'EOF'
# Kloudlite shell initialization
source /etc/kloudlite/shell-init.zsh
EOF
    chown kl:kl "$ZSHRC"
    echo ".zshrc configured"
else
    echo ".zshrc already configured"
fi

# Initialize fish configuration
mkdir -p "$KL_HOME/.config/fish"
if ! has_marker "$FISH_CONFIG" "$BASHRC_MARKER"; then
    echo "Setting up fish config..."
    cat > "$FISH_CONFIG" << 'EOF'
# Kloudlite shell initialization
source /etc/kloudlite/shell-init.fish
EOF
    chown -R kl:kl "$KL_HOME/.config/fish"
    echo "fish config configured"
else
    echo "fish config already configured"
fi

# Ensure starship config symlink exists
mkdir -p "$KL_HOME/.config"
if [ ! -L "$STARSHIP_CONFIG" ]; then
    echo "Creating starship config symlink..."
    ln -sf /etc/kloudlite/starship.toml "$STARSHIP_CONFIG"
    chown -R kl:kl "$KL_HOME/.config"
    echo "starship config linked"
fi

echo "Shell configurations initialized successfully"
