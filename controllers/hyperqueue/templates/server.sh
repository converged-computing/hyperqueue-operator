#!/bin/sh

echo "Hello, I am a server with $(hostname)"

# This script handles shared start logic
{{template "init" .}}

# We will need to copy access.json to our workdir so we have write
{{template "server-dir" .}}

sleep infinity
if [ "$@" == "" ]; then
    hq server start --access-file=./hq/access.json
else
    hq server start --access-file=./hq/access.json &
    hq submit $@
fi

{{template "exit" .}}
