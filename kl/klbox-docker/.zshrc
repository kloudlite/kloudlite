#!/bin/zsh

# If not running interactively, don't do anything
case $- in
    *i*) ;;
      *) return;;
esac

HISTCONTROL=ignoreboth

HISTSIZE=100000
SAVEHIST=100000
HISTFILE="$HOME/.zsh_history"

setopt append_history
setopt inc_append_history
setopt hist_expire_dups_first # delete duplicates first when HISTFILE size exceeds HISTSIZE
setopt hist_ignore_dups       # ignore duplicated commands history list
setopt hist_save_no_dups      # duplicate commands are not written
setopt hist_ignore_space      # ignore commands that start with space
setopt hist_verify            # show command with history expansion to user before running it
# setopt share_history          # share command history data
setopt hist_reduce_blanks     # removing blank lines from the history

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
