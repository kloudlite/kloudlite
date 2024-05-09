# ~/.profile: executed by the command interpreter for login shells.
# This file is not read by bash(1), if ~/.bash_profile or ~/.bash_login
# exists.
# see /usr/share/doc/bash/examples/startup-files for examples.
# the files are located in the bash-doc package.

# the default umask is set in /etc/profile; for setting the umask
# for ssh logins, install and configure the libpam-umask package.
#umask 022

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

source /home/kl/.nix-profile/etc/profile.d/nix.sh
prevPath=$(echo $PATH)
if [ -f $HOME/.nix-shell-args ]; then
	eval $(nix-shell --pure -p $(cat $HOME/.nix-shell-args) --run env | awk -F= '{print "export " $1 "=\"" $2 "\""}')
fi

# if [ -f $HOME/.env-vars ]; then
# 	eval $(cat $HOME/.env-vars | awk -F= '{print "export " $1 "=\"" $2 "\""}')
# fi

unset LS_COLORS
PATH="$PATH:$prevPath"

source /home/kl/.env-vars

