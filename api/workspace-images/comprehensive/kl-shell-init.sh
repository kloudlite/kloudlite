# Kloudlite workspace shell initialization for Bash
# This file is sourced by bash only

# Source /etc/environment to get Kubernetes service account env vars
# This is needed because SSH sessions don't get container environment variables automatically
if [ -f /etc/environment ]; then
    while IFS='=' read -r key value; do
        [ -z "$key" ] && continue
        case "$key" in \#*) continue ;; esac
        export "$key=$value"
    done < /etc/environment
fi

# Ensure /home/kl/.local/bin is first in PATH for user-installed npm packages
# Remove duplicates and unwanted /games directories
export PATH="/home/kl/.local/bin:$(echo "$PATH" | tr ':' '\n' | grep -v '/games' | grep -v '^/home/kl/.local/bin$' | uniq | tr '\n' ':' | sed 's/:$//')"

# Enable colored ls output
if [ -x /usr/bin/dircolors ]; then
    test -r ~/.dircolors && eval "$(dircolors -b ~/.dircolors)" || eval "$(dircolors -b)"
    alias ls='ls --color=auto'
    alias ll='ls -la --color=auto'
    alias la='ls -A --color=auto'
    alias grep='grep --color=auto'
fi

# Initialize starship prompt
export STARSHIP_CONFIG=~/.config/starship.toml
eval "$(starship init bash)"
