#!/bin/sh

echo "Hello, I am a worker with $(hostname)"

# Shared logic to install hq
{{template "init" .}}

# We will need to copy access.json to our workdir so we have write
{{template "server-dir" .}}

# Keep trying until we connect
until hq --server-dir=./hq worker start
do
    echo "Trying again to connect to main server..."
    sleep 2
done

{{template "exit" .}}