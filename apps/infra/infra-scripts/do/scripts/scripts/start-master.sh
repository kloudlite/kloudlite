#! /usr/bin/env bash

k3s server --node-name master-0 --no-deploy traefik $@ |& flog -l 10283746 /root/scripts/logs/k3s.log &
X=$!
echo "PID: $X"
disown $X
