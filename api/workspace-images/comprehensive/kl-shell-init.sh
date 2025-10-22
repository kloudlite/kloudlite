# Kloudlite workspace shell initialization for Bash
# This file is sourced by bash only

# Ensure /home/kl/.local/bin is first in PATH for user-installed npm packages
# Remove duplicates and unwanted /games directories
export PATH="/home/kl/.local/bin:$(echo "$PATH" | tr ':' '\n' | grep -v '/games' | grep -v '^/home/kl/.local/bin$' | uniq | tr '\n' ':' | sed 's/:$//')"

# Initialize starship prompt
export STARSHIP_CONFIG=~/.config/starship.toml
eval "$(starship init bash)"
