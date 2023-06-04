/*
Copyright 2023 Lawrence Livermore National Security, LLC
 (c.f. AUTHORS, NOTICE.LLNS, COPYING)

This is part of the Flux resource manager framework.
For details, see https://github.com/flux-framework.

SPDX-License-Identifier: Apache-2.0
*/

package controllers

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	ctrl "sigs.k8s.io/controller-runtime"
	jobset "sigs.k8s.io/jobset/api/v1alpha1"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"

	api "github.com/converged-computing/hyperqueue-operator/api/v1alpha1"
)

// A Hyperqueue is one or more workers plus a main server

// newHyperqueue creates a new Hyperqueue
func (r *HyperqueueReconciler) ensureHyperqueue(
	ctx context.Context,
	cluster *api.Hyperqueue,
) (ctrl.Result, error) {

	// Add entrypoint config maps
	_, result, err := r.ensureConfigMap(ctx, cluster, "entrypoint", cluster.Name+entrypointSuffix)
	if err != nil {
		return result, err
	}

	// Create headless service for the Hyperqueue cluster
	selector := map[string]string{"cluster-name": cluster.Name}
	result, err = r.exposeServices(ctx, cluster, serviceName, selector)
	if err != nil {
		return result, err
	}

	// Create the batch job that brings it all together!
	// A batchv1.Job can hold a spec for containers that use the configs we just made
	_, result, err = r.getCluster(ctx, cluster)
	if err != nil {
		return result, err
	}
	// And we re-queue so the Ready condition triggers next steps!
	return ctrl.Result{Requeue: true}, nil
}

// getExistingJob gets an existing job that matches our CRD
func (r *HyperqueueReconciler) getExistingJob(
	ctx context.Context,
	cluster *api.Hyperqueue,
) (*jobset.JobSet, error) {

	existing := &jobset.JobSet{}
	err := r.Client.Get(
		ctx,
		types.NamespacedName{
			Name:      cluster.Name,
			Namespace: cluster.Namespace,
		},
		existing,
	)
	return existing, err
}

// getCluster does an actual check if we have a jobset in the namespace
func (r *HyperqueueReconciler) getCluster(
	ctx context.Context,
	cluster *api.Hyperqueue,
) (*jobset.JobSet, ctrl.Result, error) {

	// Look for an existing job
	existing, err := r.getExistingJob(ctx, cluster)

	// Create a new job if it does not exist
	if err != nil {

		if errors.IsNotFound(err) {
			job, err := r.newJobSet(cluster)
			if err != nil {
				r.Log.Error(
					err, "Failed to create new Hyperqueue JobSet",
					"Namespace:", job.Namespace,
					"Name:", job.Name,
				)
				// If there is an error, return the existing (empty)
				return existing, ctrl.Result{}, err
			}

			r.Log.Info(
				"✨ Creating a new Hyperqueue JobSet ✨",
				"Namespace:", job.Namespace,
				"Name:", job.Name,
			)

			err = r.Client.Create(ctx, job)
			if err != nil {
				r.Log.Error(
					err,
					"Failed to create new Hyperqueue JobSet",
					"Namespace:", job.Namespace,
					"Name:", job.Name,
				)
				return existing, ctrl.Result{}, err
			}
			// Successful - return and requeue
			return job, ctrl.Result{Requeue: true}, nil

		} else if err != nil {
			r.Log.Error(err, "Failed to get Hyperqueue JobSet")
			return existing, ctrl.Result{}, err
		}

	} else {
		r.Log.Info(
			"🎉 Found existing Hyperqueue JobSet 🎉",
			"Namespace:", existing.Namespace,
			"Name:", existing.Name,
		)
	}
	return existing, ctrl.Result{}, err
}

// getConfigMap generates the config map, when does not exist
func (r *HyperqueueReconciler) getConfigMap(
	ctx context.Context,
	cluster *api.Hyperqueue,
	configName string,
	configFullName string,
) (*corev1.ConfigMap, ctrl.Result, error) {

	// Data for the config map
	data := map[string]string{}
	cm := &corev1.ConfigMap{}

	// This is currently the only config we support
	if configName == "entrypoint" {

		// Generate data for both the start-server.sh and start-worker.sh
		serverStart, err := generateScript(cluster, cluster.Spec.Server, "server")
		if err != nil {
			return cm, ctrl.Result{}, err
		}
		workerStart, err := generateScript(cluster, cluster.Spec.Worker, "worker")
		if err != nil {
			return cm, ctrl.Result{}, err
		}
		data["start-server"] = serverStart
		data["start-worker"] = workerStart
	}
	fmt.Println(data)

	// Create the config map with respective data!
	cm = &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:      configFullName,
			Namespace: cluster.Namespace,
		},
		Data: data,
	}

	// Show in the logs
	fmt.Println(cm.Data)
	ctrl.SetControllerReference(cluster, cm, r.Scheme)

	// Finally create the config map
	r.Log.Info(
		"✨ Creating Hyperqueue ConfigMap ✨",
		"Type", configName,
		"Namespace", cm.Namespace,
		"Name", cm.Name,
	)

	// Actually create it
	err := r.Create(ctx, cm)
	if err != nil {
		r.Log.Error(
			err, "❌ Failed to create Hyperqueue ConfigMap",
			"Type", configName,
			"Namespace", cm.Namespace,
			"Name", (*cm).Name,
		)
		return cm, ctrl.Result{}, err
	}

	// Successful - return and requeue
	return cm, ctrl.Result{Requeue: true}, nil
}

// ensureConfigMap ensures we've generated the read only entrypoints
func (r *HyperqueueReconciler) ensureConfigMap(
	ctx context.Context,
	cluster *api.Hyperqueue,
	configName string,
	configFullName string,
) (*corev1.ConfigMap, ctrl.Result, error) {

	// Look for the config map by name
	existing := &corev1.ConfigMap{}
	err := r.Get(
		ctx,
		types.NamespacedName{
			Name:      configFullName,
			Namespace: cluster.Namespace,
		},
		existing,
	)

	if err != nil {

		// Case 1: not found yet, and hostfile is ready (recreate)
		if errors.IsNotFound(err) {
			return r.getConfigMap(ctx, cluster, configName, configFullName)

		} else if err != nil {
			r.Log.Error(err, "Failed to get Hyperqueue ConfigMap")
			return existing, ctrl.Result{}, err
		}

	} else {
		r.Log.Info(
			"🎉 Found existing Hyperqueue ConfigMap",
			"Type", configName,
			"Namespace", existing.Namespace,
			"Name", existing.Name,
		)
	}
	return existing, ctrl.Result{}, err
}
