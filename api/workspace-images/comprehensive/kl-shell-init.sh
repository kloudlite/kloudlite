# Kloudlite workspace shell initialization
# This file is sourced by bash and zsh

export STARSHIP_CONFIG=~/.config/starship.toml
eval "$(starship init bash 2>/dev/null || starship init zsh 2>/dev/null)"
