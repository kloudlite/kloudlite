# Kloudlite workspace shell initialization
# This file is sourced by bash and zsh

# Ensure /home/kl/.local/bin is first in PATH for user-installed npm packages
# Remove duplicates and unwanted /games directories
export PATH="/home/kl/.local/bin:$(echo "$PATH" | tr ':' '\n' | grep -v '/games' | grep -v '^/home/kl/.local/bin$' | tr '\n' ':' | sed 's/:$//')"

export STARSHIP_CONFIG=~/.config/starship.toml
eval "$(starship init bash 2>/dev/null || starship init zsh 2>/dev/null)"
