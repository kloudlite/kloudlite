#!/bin/zsh

# If not running interactively, don't do anything
case $- in
    *i*) ;;
      *) return;;
esac

HISTCONTROL=ignoreboth

HISTSIZE=1000
HISTFILESIZE=2000
HISTFILE="$HOME/.zsh_history"

# zsh pure prompt setup
fpath+=($HOME/.config/zsh/pure)
autoload -U promptinit; promptinit
prompt pure
zstyle :prompt:pure:environment:nix-shell show no

# zsh aliases
if [[ -f "$HOME/.config/aliasrc" ]]; then
  source "$HOME/.config/aliasrc"
fi

# zsh syntax highlighting
ZSH_HIGHLIGHT_PATH="/usr/share/zsh/plugins/zsh-syntax-highlighting/zsh-syntax-highlighting.zsh"
if [[ -f $ZSH_HIGHLIGHT_PATH ]]; then
  source $ZSH_HIGHLIGHT_PATH
fi

# go to workspace
cd $HOME/workspace
