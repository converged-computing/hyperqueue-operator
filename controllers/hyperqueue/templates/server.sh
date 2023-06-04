#!/bin/sh

echo "Hello, I am a server with $(hostname)"

# This script handles shared start logic
{{template "init" .}}

# We will need to copy access.json to our workdir so we have write
{{template "server-dir" .}}

if [ "$@" == "" ]; then
    hq server start --server-dir ./hq
else
    hq server start --server-dir ./hq &
    hq submit $@
fi

{{template "exit" .}}
