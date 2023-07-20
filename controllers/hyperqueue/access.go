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
	"context"
	"fmt"
	"io"
	"strings"

	api "github.com/converged-computing/hyperqueue-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
)

const (
	accessContainerName = "access-generator"
	accessSuffix        = "-access"
	accessEntrypointKey = "access-generate"
	accessKey           = "access-json"
)

// getAccess brings up a one off pod to generate the access.json file
func (r *HyperqueueReconciler) getAccess(ctx context.Context, cluster *api.Hyperqueue) (string, error) {

	// This is a one time entrypoint to generate the flux curve certificate in a single pod
	_, _, err := r.generateAccessEntrypoint(ctx, cluster, cluster.Spec.Server)
	if err != nil {
		return "", err
	}

	existing := &corev1.Pod{}
	err = r.Get(ctx, types.NamespacedName{Name: cluster.Name, Namespace: cluster.Namespace}, existing)
	if err != nil {
		command := []string{"/bin/bash", fmt.Sprintf("/hyperqueue_operator/%s.sh", accessEntrypointKey)}
		pod := r.newPodCommandRunner(cluster, cluster.Spec.Server, command)
		r.Log.Info("✨ Creating a new Pod Command Runner ✨", "Namespace:", pod.Namespace, "Name:", pod.Name)

		// We are being bad and not checking if there are errors - we just want to get the certificate
		r.Create(ctx, pod)
		existing = pod
	}

	// If we get here, try to get the log output with the access.json
	access, err := r.getPodLogs(ctx, existing)

	// Split on token
	fmt.Println(access)
	parts := strings.SplitN(access, "CUT HERE", 2)
	access = parts[1]

	// Replace the generator server name with our expected server hostname
	hostname := fmt.Sprintf("%s-%s-server-0-0", cluster.Name, cluster.Name)
	access = strings.Replace(access, "hyperqueue-sample-access", hostname, 1)
	if access != "" {
		fmt.Printf("🌵 Generated access.json\n%s\n", access)
	}
	return access, err
}

// getPodLogs gets the pod logs (with the curve cert)
func (r *HyperqueueReconciler) getPodLogs(ctx context.Context, pod *corev1.Pod) (string, error) {

	// creates the clientset
	clientset, err := kubernetes.NewForConfig(r.RESTConfig)
	if err != nil {
		return "", err
	}

	// Keep developer user informed what is going on.
	r.Log.Info("Pod Logs", "Name", pod.Name)
	r.Log.Info("Pod Logs", "Container", pod.Spec.Containers[0].Name)
	opts := corev1.PodLogOptions{
		Container: pod.Spec.Containers[0].Name,
	}

	// This will fail (and need to reconcile) while container is creating, etc.
	req := clientset.CoreV1().Pods(pod.Namespace).GetLogs(pod.Name, &opts)
	podLogs, err := req.Stream(ctx)
	if err != nil {
		return "", err
	}
	defer podLogs.Close()

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, podLogs)
	if err != nil {
		return "", err
	}
	logs := buf.String()
	return logs, err
}

// generateAccessToken generate the config map entrypoint for the access.json
func (r *HyperqueueReconciler) generateAccessEntrypoint(
	ctx context.Context,
	cluster *api.Hyperqueue,
	node api.Node,
) (*corev1.ConfigMap, ctrl.Result, error) {

	existing := &corev1.ConfigMap{}
	configFullName := cluster.Name + accessSuffix
	err := r.Client.Get(
		ctx,
		types.NamespacedName{
			Name:      configFullName,
			Namespace: cluster.Namespace,
		},
		existing,
	)

	if err != nil {

		// Case 1: not found yet, so we generate the pod
		if errors.IsNotFound(err) {

			data := map[string]string{}
			accessEntrypoint, err := generateScript(cluster, node, accessTemplate)
			if err != nil {
				return existing, ctrl.Result{}, err
			}
			data[accessEntrypointKey] = accessEntrypoint
			cm := r.createConfigMap(cluster, configFullName, data)
			err = r.Client.Create(ctx, cm)
			if err != nil {
				r.Log.Error(
					err, "❌ Failed to create access.json Pod Generator Entrypoint",
					"Namespace", cm.Namespace,
					"Name", cm.Name,
				)
				return existing, ctrl.Result{}, err
			}
			// Successful - return and requeue
			return cm, ctrl.Result{Requeue: true}, nil

		} else if err != nil {
			r.Log.Error(err, "Failed to get access.json Pod Generator Entrypoint")
			return existing, ctrl.Result{}, err
		}
	} else {
		r.Log.Info(
			"🎉 Found existing access.json Pod Generator Entrypoint",
			"Namespace", existing.Namespace,
			"Name", existing.Name,
		)
	}
	return existing, ctrl.Result{}, err
}

// createPersistentVolume creates a volume in /tmp, which doesn't seem to choke
func (r *HyperqueueReconciler) newPodCommandRunner(
	cluster *api.Hyperqueue,
	node api.Node,
	command []string,
) *corev1.Pod {

	makeExecutable := int32(0777)
	pullPolicy := corev1.PullIfNotPresent
	if node.PullAlways {
		pullPolicy = corev1.PullAlways
	}

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cluster.Name + accessSuffix,
			Namespace: cluster.Namespace,
		},
		Spec: corev1.PodSpec{
			RestartPolicy: corev1.RestartPolicyOnFailure,
			Volumes: []corev1.Volume{{
				Name: cluster.Name + accessSuffix,
				VolumeSource: corev1.VolumeSource{
					ConfigMap: &corev1.ConfigMapVolumeSource{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: cluster.Name + accessSuffix,
						},
						Items: []corev1.KeyToPath{{
							Key:  accessEntrypointKey,
							Path: fmt.Sprintf("%s.sh", accessEntrypointKey),
							Mode: &makeExecutable,
						}},
					},
				},
			}},
			Containers: []corev1.Container{{
				Name:            accessContainerName,
				Image:           node.Image,
				ImagePullPolicy: pullPolicy,
				WorkingDir:      node.WorkingDir,
				Stdin:           true,
				VolumeMounts: []corev1.VolumeMount{
					{
						Name:      cluster.Name + accessSuffix,
						MountPath: "/hyperqueue_operator/",
						ReadOnly:  true,
					}},
				TTY:     true,
				Command: command,
			}},
		},
	}
	// Do we have pull secrets?
	if node.PullSecret != "" {
		pod.Spec.ImagePullSecrets = []corev1.LocalObjectReference{
			{Name: node.PullSecret},
		}
	}

	ctrl.SetControllerReference(cluster, pod, r.Scheme)
	return pod
}
