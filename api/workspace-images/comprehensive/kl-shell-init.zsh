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

# Configure npm to install global packages in user directory
# This allows npm install -g to work without sudo
export NPM_CONFIG_PREFIX="$HOME/.local"

# Initialize starship prompt
export STARSHIP_CONFIG=~/.config/starship.toml
eval "$(starship init zsh)"
