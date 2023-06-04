#!/bin/sh

# This script handles shared start logic
{{template "init" .}}

# Start the server
hq server start &

if [ "$@" == "" ]; then
    hq server start
else
    hq server start &
    hq server submit $@
fi

{{template "exit" .}}
