#!/bin/bash
# Kloudlite Workspace Welcome Message

cat << 'EOF'

  ╦╔═╦  ╔═╗╦ ╦╔╦╗╦  ╦╔╦╗╔═╗
  ╠╩╗║  ║ ║║ ║ ║║║  ║ ║ ║╣
  ╩ ╩╩═╝╚═╝╚═╝═╩╝╩═╝╩ ╩ ╚═╝

  Welcome to your Kloudlite Workspace

EOF

# Show workspace info if available
if [ -n "$WORKSPACE_NAME" ]; then
    echo "  Workspace: $WORKSPACE_NAME"
fi

# Show available tools
echo "  Shell: fish (with starship prompt)"
echo "  Tools: code-server (port 8080), ttyd (port 7681)"
echo ""
echo "  Get started:"
echo "    • Run 'claude' to start Claude Code"
echo "    • Your workspace directory: /workspace"
echo "    • Shared files: /home/kl/workspaces"
echo ""
