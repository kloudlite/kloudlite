#!/bin/bash

# ~/.profile: executed by the command interpreter for login shells.
# This file is not read by bash(1), if ~/.bash_profile or ~/.bash_login
# exists.
# see /usr/share/doc/bash/examples/startup-files for examples.
# the files are located in the bash-doc package.

# the default umask is set in /etc/profile; for setting the umask
# for ssh logins, install and configure the libpam-umask package.
#umask 022

KL_DEVBOX_PATH=/home/kl/.kl/devbox

mkdir -p $KL_DEVBOX_PATH

cp /tmp/sample.json $KL_DEVBOX_PATH/devbox.json

if [ -f "/home/kl/workspace/kl.lock" ]; then
    cp /home/kl/workspace/kl.lock $KL_DEVBOX_PATH/devbox.lock
fi

if [ ! -f "/tmp/devbox.sh" ]; then
    cd "$KL_DEVBOX_PATH" 
    devbox shellenv > /tmp/devbox.sh
fi

source /tmp/devbox.sh
cd $HOME

if [ -f "$KL_DEVBOX_PATH/devbox.lock" ]; then
  cp "$KL_DEVBOX_PATH/devbox.lock" /home/kl/workspace/kl.lock
fi

sudo /mounter --conf /tmp/sample.json

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


if [ -f "/home/kl/.kl/global-profile" ]; then
  source /home/kl/.kl/global-profile
fi

cd /home/kl/workspace
