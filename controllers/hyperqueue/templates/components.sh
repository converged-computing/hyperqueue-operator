#!/bin/sh

# Shared components for the broker and worker template
{{define "init"}}

# Initialization commands
{{ .Node.Commands.Init}} > /dev/null 2>&1

which wget > /dev/null 2>&1 || (echo "Please install wget"; exit);

function download() {
    wget https://github.com/It4innovations/hyperqueue/releases/download/v{{ .Spec.HyperqueueVersion }}/hq-v{{ .Spec.HyperqueueVersion }}-linux-x64.tar.gz
    tar -xvzf hq-v{{ .Spec.HyperqueueVersion }}-linux-x64.tar.gz
    mv hq /usr/bin/hq
}

# If hyperqueue isn't installed, install it
# which hq > /dev/null 2>&1 || (download > /dev/null 2>&1);
# Download development version for now

# The working directory should be set by the CRD or the container
workdir=${PWD}

# And if we are using fusefs / object storage, ensure we can see contents
mkdir -p ${workdir}

# End init logic
{{end}}

{{define "exit"}}
{{ if .Spec.Interactive }}sleep infinity{{ end }}
{{ end }}

{{define "server-dir"}}
mkdir -p ./hq
cp /hyperqueue_operator/access.json ./hq/access.json
{{ end }}