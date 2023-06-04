#!/bin/sh

echo "Hello, I am a server with $(hostname)"

# This script handles shared start logic
{{template "init" .}}

# Start the server
hq server start &

if [ "$@" == "" ]; then
    hq server start
else
    hq server start &
    hq submit $@
fi

{{template "exit" .}}
