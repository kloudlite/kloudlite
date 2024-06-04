#!/bin/bash

(tail -f /tmp/stdout.log) &
pid=$!
(tail -f /tmp/stderr.log) &
pid="$pid $1"

trap "eval kill -9 $pid" EXIT TERM
/start.sh $@ > /tmp/stdout.log 2> /tmp/stderr.log
