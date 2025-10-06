#!/bin/bash
set -e

# Source Nix environment
if [ -f /home/workspace/.nix-profile/etc/profile.d/nix.sh ]; then
    . /home/workspace/.nix-profile/etc/profile.d/nix.sh
fi

echo "==> Workspace environment initializing..."

# Check if workspace-specific packages configuration exists
WORKSPACE_PACKAGES_FILE="${WORKSPACE_PACKAGES_FILE:-/workspace/.kloudlite/packages.yaml}"

if [ -f "$WORKSPACE_PACKAGES_FILE" ]; then
    echo "==> Found workspace packages configuration: $WORKSPACE_PACKAGES_FILE"
    echo "==> Installing packages..."

    # Parse YAML and install packages
    # Read the packages list from YAML
    PACKAGES=$(yq eval '.packages[]' "$WORKSPACE_PACKAGES_FILE" 2>/dev/null || echo "")

    if [ -n "$PACKAGES" ]; then
        echo "==> Installing Nix packages..."
        for pkg in $PACKAGES; do
            echo "  - Installing: $pkg"
            nix-env -iA "nixpkgs.$pkg" 2>&1 || {
                echo "  WARNING: Failed to install $pkg, trying alternative method..."
                nix-env -i "$pkg" 2>&1 || echo "  ERROR: Could not install $pkg"
            }
        done
        echo "==> Package installation complete"
    else
        echo "==> No packages found in configuration file"
    fi
else
    echo "==> No workspace packages configuration found at $WORKSPACE_PACKAGES_FILE"
    echo "==> Using default packages"
fi

# Run startup script if provided
if [ -n "$STARTUP_SCRIPT" ] && [ -f "$STARTUP_SCRIPT" ]; then
    echo "==> Running startup script: $STARTUP_SCRIPT"
    bash "$STARTUP_SCRIPT" || {
        echo "WARNING: Startup script failed"
    }
fi

echo "==> Workspace environment ready!"

# Execute the command passed to the container
exec "$@"
