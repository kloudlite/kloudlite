# Kloudlite workspace shell initialization for fish
# This file is sourced by fish shell

# Ensure /home/kl/.local/bin is first in PATH for user-installed npm packages
# Remove duplicates and unwanted /games directories
set -x PATH /home/kl/.local/bin (string split : $PATH | string match -v -r '/games' | string match -v /home/kl/.local/bin)

# Disable fish greeting
set -g fish_greeting

# Fix terminal type warnings
set -x TERM xterm-256color

# Set starship config and initialize
set -x STARSHIP_CONFIG ~/.config/starship.toml
starship init fish | source
