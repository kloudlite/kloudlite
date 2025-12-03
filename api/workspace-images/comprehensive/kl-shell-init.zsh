# Kloudlite workspace shell initialization for Zsh
# This file is sourced by zsh only

# Source /etc/environment to get Kubernetes service account env vars
# This is needed because SSH sessions don't get container environment variables automatically
if [[ -f /etc/environment ]]; then
    while IFS='=' read -r key value; do
        [[ -z "$key" || "$key" == \#* ]] && continue
        export "$key=$value"
    done < /etc/environment
fi

# Ensure /home/kl/.local/bin is first in PATH for user-installed npm packages
# Remove duplicates and unwanted /games directories
typeset -U path
path=(/home/kl/.local/bin ${path[@]:#*/games})

# Enable colored ls output
if [[ -x /usr/bin/dircolors ]]; then
    test -r ~/.dircolors && eval "$(dircolors -b ~/.dircolors)" || eval "$(dircolors -b)"
    alias ls='ls --color=auto'
    alias ll='ls -la --color=auto'
    alias la='ls -A --color=auto'
    alias grep='grep --color=auto'
fi

# Initialize zsh completions
autoload -Uz compinit && compinit

# Initialize starship prompt
export STARSHIP_CONFIG=~/.config/starship.toml
eval "$(starship init zsh)"

# Initialize kl shell completions
if (( $+commands[kl] )); then
    source <(kl completion zsh)
fi
