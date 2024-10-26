#!/bin/bash
# shellcheck source=/dev/null

# ~/.profile: executed by the command interpreter for login shells.
# This file is not read by bash(1), if ~/.bash_profile or ~/.bash_login
# exists.
# see /usr/share/doc/bash/examples/startup-files for examples.
# the files are located in the bash-doc package.

# the default umask is set in /etc/profile; for setting the umask
# for ssh logins, install and configure the libpam-umask package.
#umask 022

# source $HOME/.nix-profile/etc/profile.d/nix.sh
export PATH=$PATH:/home/kl/.nix-profile/bin

KL_LOCK_PATH=/home/kl/workspace/kl.lock
[ -f /kl-tmp/env ] && echo "sourcing /kl-tmp/env" && source /kl-tmp/env

# KL_DEVBOX_PATH=/kl-tmp/devbox
# KL_DEVBOX_JSON_PATH=$KL_DEVBOX_PATH/devbox.json
# KL_DEVBOX_LOCK_PATH=$KL_DEVBOX_PATH/devbox.lock
# KL_DEVBOX_ENV_PATH=/kl-tmp/devbox/devbox-env.sh

# mkdir -p $KL_DEVBOX_PATH

# if [ -f "$KL_LOCK_PATH" ]; then
#     cp $KL_LOCK_PATH $KL_DEVBOX_LOCK_PATH
# fi
#
# if [ ! -f "$KL_DEVBOX_ENV_PATH" ]; then
#     cd  $KL_DEVBOX_PATH || return
#     devbox shellenv > $KL_DEVBOX_ENV_PATH
# fi

# source "$KL_DEVBOX_ENV_PATH"
#
# cd "$HOME" || return
#
# if [ -f "$KL_DEVBOX_LOCK_PATH" ]; then
#   cp $KL_DEVBOX_LOCK_PATH $KL_LOCK_PATH
# fi
#
# sudo /mounter --conf $KL_DEVBOX_JSON_PATH
#
# if running bash
if [ -n "$BASH_VERSION" ]; then
	# include .bashrc if it exists
	if [ -f "$HOME/.bashrc" ]; then
		. "$HOME/.bashrc"
	fi
fi

# set PATH so it includes user's private bin if it exists
if [ -d "$HOME/bin" ]; then
	PATH="$HOME/bin:$PATH"
fi


# set PATH so it includes user's private bin if it exists
if [ -d "$HOME/.local/bin" ]; then
	PATH="$HOME/.local/bin:$PATH"
fi

if [ -f "/kl-tmp/global-profile" ]; then
  source /kl-tmp/global-profile
fi

cd /home/kl/workspace || return
