#!/bin/bash

if ping -c 1 100.64.0.1 > /dev/null 2>&1
then
    exit 0
else
    exit 1
fi
