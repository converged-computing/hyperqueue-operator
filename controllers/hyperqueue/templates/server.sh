#!/bin/sh

# This script handles start logic for the broker
{{template "init" .}}

# Start the server
hq server start &

if [ "$@" == "" ]; then
    hq server start
else
    hq server start &
    hq server submit $@
fi
