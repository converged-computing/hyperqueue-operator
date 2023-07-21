/*
Copyright 2023 Lawrence Livermore National Security, LLC
 (c.f. AUTHORS, NOTICE.LLNS, COPYING)

This is part of the Flux resource manager framework.
For details, see https://github.com/flux-framework.

SPDX-License-Identifier: Apache-2.0
*/

package controllers

import (
	api "github.com/converged-computing/hyperqueue-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

const (
	entrypointSuffix   = "-entrypoint"
	accessVolumeSuffix = "-access-json"
)

// GetVolumeMounts returns read only volume for entrypoint scripts, etc.
func getVolumeMounts(cluster *api.Hyperqueue) []corev1.VolumeMount {
	mounts := []corev1.VolumeMount{
		{
			Name:      cluster.Name + entrypointSuffix,
			MountPath: "/hyperqueue_operator/",
			ReadOnly:  true,
		},
	}
	return mounts
}

// getVolumes for the Indexed Jobs
func getVolumes(cluster *api.Hyperqueue) []corev1.Volume {

	// Runner start scripts
	makeExecutable := int32(0777)

	// Each of the server and nodes are given the entrypoint scripts
	// Although they won't both use them, this makes debugging easier
	runnerScripts := []corev1.KeyToPath{
		{
			Key:  "start-server",
			Path: "start-server.sh",
			Mode: &makeExecutable,
		},
		{
			Key:  "start-worker",
			Path: "start-worker.sh",
			Mode: &makeExecutable,
		},
		{
			Key:  accessKey,
			Path: "access.json",
		},
	}

	volumes := []corev1.Volume{
		{
			Name: cluster.Name + entrypointSuffix,
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{

					// Namespace based on the cluster
					LocalObjectReference: corev1.LocalObjectReference{
						Name: cluster.Name + entrypointSuffix,
					},
					// /hyperqueue_operator/start-worker.sh
					// /hyperqueue_operator/start-server.sh
					// /hyperqueue_operator/access.json
					Items: runnerScripts,
				},
			},
		},
	}

	// Add volumes that already exist (not created by the Flux Operator)
	// These are unique names and path/claim names across containers
	// This can be a claim, secret, or config map
	existingVolumes := getExistingVolumes(cluster.ExistingContainerVolumes())
	volumes = append(volumes, existingVolumes...)
	return volumes
}

// Get Existing volumes for the cluster
func getExistingVolumes(existing map[string]api.ExistingVolume) []corev1.Volume {
	volumes := []corev1.Volume{}
	for volumeName, volumeMeta := range existing {

		var newVolume corev1.Volume
		if volumeMeta.SecretName != "" {
			newVolume = corev1.Volume{
				Name: volumeName,
				VolumeSource: corev1.VolumeSource{
					Secret: &corev1.SecretVolumeSource{
						SecretName: volumeMeta.SecretName,
					},
				},
			}

		} else if volumeMeta.ConfigMapName != "" {

			// Prepare items as key to path
			items := []corev1.KeyToPath{}
			for key, path := range volumeMeta.Items {
				newItem := corev1.KeyToPath{
					Key:  key,
					Path: path,
				}
				items = append(items, newItem)
			}

			// This is a config map volume with items
			newVolume = corev1.Volume{
				Name: volumeMeta.ConfigMapName,
				VolumeSource: corev1.VolumeSource{
					ConfigMap: &corev1.ConfigMapVolumeSource{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: volumeMeta.ConfigMapName,
						},
						Items: items,
					},
				},
			}

		} else {

			// Fall back to persistent volume claim
			newVolume = corev1.Volume{
				Name: volumeName,
				VolumeSource: corev1.VolumeSource{
					PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
						ClaimName: volumeMeta.ClaimName,
					},
				},
			}
		}
		volumes = append(volumes, newVolume)
	}
	return volumes
}
