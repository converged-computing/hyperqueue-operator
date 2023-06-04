#!/bin/sh

# This script handles shared start logic
{{template "init" .}}

# The entire purpose of this script is to start a server to generate an access.json
# This should start and exit cleanly
mkdir -p /tmp/access
hq server start --server-dir /tmp/access &
sleep 2
echo "CUT HERE"
cat /tmp/access/001/access.json