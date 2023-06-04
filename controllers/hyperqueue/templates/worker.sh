#!/bin/sh

echo "Hello, I am a worker with $(hostname)"

# Shared logic to install hq
{{template "init" .}}

hq worker start --server-dir ./hq

{{template "exit" .}}