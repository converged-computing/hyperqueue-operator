#!/bin/sh

# This script handles shared start logic
{{template "init" .}}

# The entire purpose of this script is to start a server to generate an access.json
# This should start and exit cleanly
# For now use the same port for server and worker, not sure why should be different?
mkdir -p /tmp/access
hq server generate-access operator-access.json --client-port={{ .Spec.Server.Port }} --worker-port={{ .Spec.Worker.Port }} --host {{ .ClusterName }}-server-0-0.{{ .Spec.ServiceName }}.{{ .Namespace }}.svc.cluster.local

sleep 2
echo "CUT HERE"
cat operator-access.json
