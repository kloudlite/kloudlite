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
SSH_CONFIG="$KL_HOME/.ssh/config"

# Marker to identify our configuration
BASHRC_MARKER="# Kloudlite shell initialization"
BASH_PROFILE_MARKER="# Source .bashrc for login shells"
TERM_FIX_MARKER="# Fix TERM for SSH login"
SSH_CONFIG_MARKER="# Kloudlite SSH configuration"

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

# Initialize SSH client configuration
mkdir -p "$KL_HOME/.ssh"
chmod 700 "$KL_HOME/.ssh"
if ! has_marker "$SSH_CONFIG" "$SSH_CONFIG_MARKER"; then
    echo "Setting up SSH client config..."
    cat > "$SSH_CONFIG" << 'EOF'
# Kloudlite SSH configuration
# Use the workspace's SSH key for all connections
Host *
    IdentityFile /var/lib/kloudlite/ssh-config/ssh_host_rsa_key
    StrictHostKeyChecking no
    UserKnownHostsFile /dev/null
EOF
    chmod 600 "$SSH_CONFIG"
    chown -R kl:kl "$KL_HOME/.ssh"
    echo "SSH client config configured"
else
    echo "SSH client config already configured"
fi

# Initialize Claude Code MCP configuration in ~/.claude.json
CLAUDE_JSON="$KL_HOME/.claude.json"
CLAUDE_MD="$KL_HOME/CLAUDE.md"
CLAUDE_MCP_MARKER='"kloudlite"'

# Add MCP config to ~/.claude.json (merge if file exists)
if ! grep -q "$CLAUDE_MCP_MARKER" "$CLAUDE_JSON" 2>/dev/null; then
    echo "Setting up Claude MCP configuration..."
    if [ -f "$CLAUDE_JSON" ]; then
        # File exists - use jq to merge if available, otherwise use Python
        if command -v jq &> /dev/null; then
            jq '. + {"mcpServers": {"kloudlite": {"command": "kl", "args": ["mcp"]}}}' "$CLAUDE_JSON" > "$CLAUDE_JSON.tmp" && mv "$CLAUDE_JSON.tmp" "$CLAUDE_JSON"
        elif command -v python3 &> /dev/null; then
            python3 -c "
import json
with open('$CLAUDE_JSON', 'r') as f:
    data = json.load(f)
data['mcpServers'] = {'kloudlite': {'command': 'kl', 'args': ['mcp']}}
with open('$CLAUDE_JSON', 'w') as f:
    json.dump(data, f, indent=2)
"
        else
            echo "Warning: Could not merge MCP config (jq or python3 required)"
        fi
    else
        # Create new file
        cat > "$CLAUDE_JSON" << 'EOF'
{
  "mcpServers": {
    "kloudlite": {
      "command": "kl",
      "args": ["mcp"]
    }
  }
}
EOF
    fi
    chown kl:kl "$CLAUDE_JSON"
    echo "Claude MCP configuration added"
else
    echo "Claude MCP configuration already exists"
fi

# Copy CLAUDE.md to home directory if not present
if [ -f /etc/kloudlite/CLAUDE.md ] && [ ! -f "$CLAUDE_MD" ]; then
    echo "Copying CLAUDE.md to home directory..."
    cp /etc/kloudlite/CLAUDE.md "$CLAUDE_MD"
    chown kl:kl "$CLAUDE_MD"
    echo "CLAUDE.md copied"
fi

# Initialize Codex CLI MCP configuration in ~/.codex/config.toml
CODEX_DIR="$KL_HOME/.codex"
CODEX_CONFIG="$CODEX_DIR/config.toml"
CODEX_MCP_MARKER="mcp_servers.kloudlite"

mkdir -p "$CODEX_DIR"
if ! grep -q "$CODEX_MCP_MARKER" "$CODEX_CONFIG" 2>/dev/null; then
    echo "Setting up Codex MCP configuration..."
    cat >> "$CODEX_CONFIG" << 'EOF'

# Kloudlite MCP server for workspace management
[mcp_servers.kloudlite]
command = "kl"
args = ["mcp"]
EOF
    chown -R kl:kl "$CODEX_DIR"
    echo "Codex MCP configuration added"
else
    echo "Codex MCP configuration already exists"
fi

# Initialize OpenCode MCP configuration in ~/.config/opencode/opencode.json
OPENCODE_DIR="$KL_HOME/.config/opencode"
OPENCODE_CONFIG="$OPENCODE_DIR/opencode.json"
OPENCODE_MCP_MARKER='"kloudlite"'

mkdir -p "$OPENCODE_DIR"
if [ ! -f "$OPENCODE_CONFIG" ] || ! grep -q "$OPENCODE_MCP_MARKER" "$OPENCODE_CONFIG" 2>/dev/null; then
    echo "Setting up OpenCode MCP configuration..."
    if [ -f "$OPENCODE_CONFIG" ]; then
        # File exists - merge with Python
        if command -v python3 &> /dev/null; then
            python3 -c "
import json
import os
config_path = '$OPENCODE_CONFIG'
try:
    with open(config_path, 'r') as f:
        data = json.load(f)
except:
    data = {}
if 'mcp' not in data:
    data['mcp'] = {}
data['mcp']['kloudlite'] = {
    'type': 'local',
    'command': ['kl', 'mcp'],
    'enabled': True
}
with open(config_path, 'w') as f:
    json.dump(data, f, indent=2)
"
        fi
    else
        # Create new file
        cat > "$OPENCODE_CONFIG" << 'EOF'
{
  "mcp": {
    "kloudlite": {
      "type": "local",
      "command": ["kl", "mcp"],
      "enabled": true
    }
  }
}
EOF
    fi
    chown -R kl:kl "$OPENCODE_DIR"
    echo "OpenCode MCP configuration added"
else
    echo "OpenCode MCP configuration already exists"
fi

echo "Shell configurations initialized successfully"
