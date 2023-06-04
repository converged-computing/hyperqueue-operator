/*
Copyright 2023 Lawrence Livermore National Security, LLC
 (c.f. AUTHORS, NOTICE.LLNS, COPYING)

This is part of the Flux resource manager framework.
For details, see https://github.com/flux-framework.

SPDX-License-Identifier: Apache-2.0
*/

package controllers

import (
	"bytes"
	"text/template"

	api "github.com/converged-computing/hyperqueue-operator/api/v1alpha1"

	_ "embed"
)

//go:embed templates/server.sh
var startServerTemplate string

//go:embed templates/worker.sh
var startWorkerTemplate string

// NodeTemplate populates a node entrypoint
type NodeTemplate struct {
	Node api.Node
	Spec api.HyperqueueSpec
}

// generateWorkerScript generates the main script to start everything up!
func generateScript(cluster *api.Hyperqueue, node api.Node, tag string) (string, error) {
	return "", nil
}

// generateWorkerScript generates the main script to start everything up!
func generateWorkerScript(cluster *api.Hyperqueue, node api.Node, tag string) (string, error) {

	nt := NodeTemplate{
		Node: node,
		Spec: cluster.Spec,
	}

	templateScript := startServerTemplate
	if tag == "worker" {
		templateScript = startWorkerTemplate
	}
	t, err := template.New(tag).Parse(templateScript)
	if err != nil {
		return "", err
	}

	var output bytes.Buffer
	if err := t.Execute(&output, nt); err != nil {
		return "", err
	}
	return output.String(), nil
}
