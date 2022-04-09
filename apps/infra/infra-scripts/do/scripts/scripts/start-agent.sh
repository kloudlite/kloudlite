#!/bin/sh

k3s agent $@ |& flog -l 10283746 /root/scripts/logs/k3s.log &
