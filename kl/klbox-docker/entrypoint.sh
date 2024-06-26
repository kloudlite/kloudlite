#! /usr/bin/env bash

/start.sh


export SSH_PORT=$SSH_PORT
trap - EXIT SIGTERM SIGINT
echo "kloudlite-entrypoint:SETUP_COMPLETE"

#/track-changes.sh "$KL_HASH_FILE" "echo kl-hash-file changed, exiting ...; sudo pkill -9 sshd" &

sudo /usr/sbin/sshd -D -p "$SSH_PORT"
