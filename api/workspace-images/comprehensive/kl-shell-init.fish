# Kloudlite workspace shell initialization for fish
# This file is sourced by fish shell

# Source /etc/environment to get Kubernetes service account env vars
# This is needed because SSH sessions don't get container environment variables automatically
if test -f /etc/environment
    while read -l line
        # Skip empty lines and comments
        test -z "$line"; and continue
        string match -q '#*' -- $line; and continue
        # Split on first = and export
        set -l parts (string split -m 1 '=' -- $line)
        if test (count $parts) -eq 2
            set -gx $parts[1] $parts[2]
        end
    end < /etc/environment
end

# Ensure /home/kl/.local/bin is first in PATH for user-installed npm packages
# Remove duplicates and unwanted /games directories
set -x PATH /home/kl/.local/bin (string split : $PATH | string match -v -r '/games' | string match -v /home/kl/.local/bin)

# Disable fish greeting
set -g fish_greeting

# Enable colored ls output
alias ls='ls --color=auto'
alias ll='ls -la --color=auto'
alias la='ls -A --color=auto'
alias grep='grep --color=auto'

# Fix terminal type warnings
set -x TERM xterm-256color

# Set starship config and initialize
set -x STARSHIP_CONFIG ~/.config/starship.toml
starship init fish | source
