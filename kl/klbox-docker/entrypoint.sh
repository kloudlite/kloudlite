#! /usr/bin/env bash

/start.sh


export SSH_PORT=$SSH_PORT
sudo /usr/sbin/sshd -D -p "$SSH_PORT"
