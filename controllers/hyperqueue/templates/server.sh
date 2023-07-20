#!/bin/sh

echo "Hello, I am a server with $(hostname)"

# This script handles shared start logic
{{template "init" .}}

# We will need to copy access.json to our workdir so we have write
{{template "server-dir" .}}

if [ "$@" == "" ]; then
    hq server start --access-file=./hq/access.json
else
    echo "Found extra command $@"
    hq server start --access-file=./hq/access.json &

    # Sleep just for a brief time to give time for workers
    # wait for job to finish. Note --cpus is an argument too
    sleep 2
    echo "hq submit --wait --name {{.Spec.Job.Name}} --nodes {{.Spec.Job.Nodes }} {{if .Spec.Job.Log}}--log {{.Spec.Job.Log}}{{end}} $@"
    hq submit --wait --name {{.Spec.Job.Name}} --nodes {{.Spec.Job.Nodes }} {{if .Spec.Job.Log}}--log {{.Spec.Job.Log}}{{end}} $@
    {{if .Spec.Job.Log}}cat {{.Spec.Job.Log}}{{end}}
fi

{{template "exit" .}}
