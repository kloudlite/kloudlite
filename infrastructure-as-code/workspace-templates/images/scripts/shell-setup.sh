#!/usr/bin/env bash

kloudlite_config_dir="$HOME/.kloudlite"
mkdir -p "$kloudlite_config_dir"

if [ ! -f "$kloudlite_config_dir/zshrc" ]; then
  cat >"$kloudlite_config_dir/zshrc" <<'EOF'
export CLICOLOR=1
export LSCOLORS=ExGxBxDxCxEgEdxbxgxcxd

# common aliases
alias ll="ls -alG"
alias ls='ls --color'

export EDITOR=nano
export PAGER=less

HISTCONTROL=ignoreboth

HISTSIZE=100000
SAVEHIST=100000
HISTFILE="$HOME/.zsh_history"

setopt append_history
setopt inc_append_history
setopt hist_expire_dups_first # delete duplicates first
setopt hist_ignore_dups       # ignore duplicated commands history list
setopt hist_save_no_dups      # duplicate commands are not written
setopt hist_ignore_space      # ignore commands that start with space
setopt hist_verify            # show command with history expansion to user before running it
# setopt share_history          # share command history data

setopt hist_reduce_blanks     # removing blank lines from the history


ZSH_AUTO_SUGGESTIONS_PATH="$HOME/.config/zsh/zsh-autosuggestions/zsh-autosuggestions.zsh"
if [[ -f $ZSH_AUTO_SUGGESTIONS_PATH ]]; then
  source $ZSH_AUTO_SUGGESTIONS_PATH
fi

# zsh syntax highlighting
ZSH_HIGHLIGHT_PATH="$HOME/.config/zsh/zsh-syntax-highlighting/zsh-syntax-highlighting.zsh"
if [[ -f $ZSH_HIGHLIGHT_PATH ]]; then
  source $ZSH_HIGHLIGHT_PATH
fi

export PATH=$PATH:$HOME/.local/bin
if [[ -f "$HOME/.nix-profile/etc/profile.d/nix.sh" ]]; then
  source $HOME/.nix-profile/etc/profile.d/nix.sh
fi

mkdir -p $KL_WORKSPACE
cd $KL_WORKSPACE

precmd() {
 export P=$(kl checkchanges)
}

reload () {
  echo "Reloading..."
  kubectl rollout restart deployment $DEPLOYMENT_NAME >/dev/null 2>&1
  kubectl rollout status deployment $DEPLOYMENT_NAME --watch >/dev/null 2>&1
}

eval "$(starship init zsh)"
EOF
fi

if [ ! -f "$kloudlite_config_dir/starship.toml" ]; then
  cat >"$kloudlite_config_dir/starship.toml" <<'EOF'
format = """
$directory\
$git_branch\
$git_state\
$git_status\
$line_break\
$env_var\
$python\
$character"""

[directory]
style = "blue"

[character]
success_symbol = "[❯](purple)"
error_symbol = "[❯](red)"
vimcmd_symbol = "[❮](green)"

[git_branch]
format = "[$branch]($style)"
style = "bright-black"

[git_status]
format = "[[(*$conflicted$untracked$modified$staged$renamed$deleted)](218) ($ahead_behind$stashed)]($style)"
style = "cyan"
conflicted = ""
untracked = ""
modified = ""
staged = ""
renamed = ""
deleted = ""
stashed = "≡"

[git_state]
format = '\([$state( $progress_current/$progress_total)]($style)\) '
style = "bright-black"

[env_var]
variable = 'P'
format = "[$env_value]($style)"

[python]
format = "[$virtualenv]($style) "
style = "bright-black"

EOF
fi

if [ ! -f "$HOME/.zshrc" ]; then
  touch "$HOME/.zshrc"
fi

if ! grep "source $kloudlite_config_dir/zshrc" <"$HOME/.zshrc"; then
  echo "source $kloudlite_config_dir/zshrc" >>"$HOME/.zshrc"
fi

if [ ! -e "$HOME/.config/starship.toml" ]; then
  ln -sf "$kloudlite_config_dir/starship.toml" "$HOME/.config/starship.toml"
fi
