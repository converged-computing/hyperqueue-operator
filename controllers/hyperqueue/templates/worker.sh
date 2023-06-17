#!/bin/sh

echo "Hello, I am a worker with $(hostname)"

# Shared logic to install hq
{{template "init" .}}

# We will need to copy access.json to our workdir so we have write
{{template "server-dir" .}}
sleep infinity
hq --server-dir=./hq worker start
{{template "exit" .}}