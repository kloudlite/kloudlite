# Kloudlite workspace shell initialization for Zsh
# This file is sourced by zsh only

# Ensure /home/kl/.local/bin is first in PATH for user-installed npm packages
# Remove duplicates and unwanted /games directories
typeset -U path
path=(/home/kl/.local/bin ${path[@]:#*/games})

# Initialize starship prompt
export STARSHIP_CONFIG=~/.config/starship.toml
eval "$(starship init zsh)"
