#! /usr/bin/env bash

if [ $(id -u) -ne 0 ]; then
	echo "This script must be run as root"
	exit 1
fi

HOST_DOCKER_SOCKET=/var/run/host-docker.sock

rm -f /var/run/docker.sock
socat UNIX-LISTEN:/var/run/docker.sock,fork,reuseaddr UNIX-CONNECT:$HOST_DOCKER_SOCKET &
pid=$!

while true; do
	if [ -x /var/run/docker.sock ]; then
		chown kl /var/run/docker.sock
		break
	fi
	sleep 1
done

wait $pid