# Kloudlite workspace shell initialization for fish
# This file is sourced by fish shell

# Disable fish greeting
set -g fish_greeting

# Fix terminal type warnings
set -x TERM xterm-256color

# Set starship config and initialize
set -x STARSHIP_CONFIG ~/.config/starship.toml
starship init fish | source
