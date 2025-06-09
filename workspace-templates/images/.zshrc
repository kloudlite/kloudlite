export CLICOLOR=1
export LSCOLORS=ExGxBxDxCxEgEdxbxgxcxd
alias ll="ls -alG"
alias ls='ls --color'

eval "$(starship init zsh)"

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


ZSH_AUTO_SUGGESTIONS_PATH="/home/kl/.config/zsh/zsh-autosuggestions/zsh-autosuggestions.zsh"
if [[ -f $ZSH_AUTO_SUGGESTIONS_PATH ]]; then
  source $ZSH_AUTO_SUGGESTIONS_PATH
fi

# zsh syntax highlighting
ZSH_HIGHLIGHT_PATH="/home/kl/.config/zsh/zsh-syntax-highlighting/zsh-syntax-highlighting.zsh"
if [[ -f $ZSH_HIGHLIGHT_PATH ]]; then
  source $ZSH_HIGHLIGHT_PATH
fi

export PATH=$PATH:$HOME/.local/bin
source /home/kl/.nix-profile/etc/profile.d/nix.sh
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
