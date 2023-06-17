/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	"reflect"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// HyperqueueSpec defines the desired state of Hyperqueue
type HyperqueueSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Server is the main server to run hyperqueue
	Server Node `json:"server"`

	// Name for the cluster service
	//+optional
	ServiceName string `json:"serviceName"`

	// Worker is the worker node spec
	// Defaults to be same spec as the server
	//+optional
	Worker Node `json:"worker"`

	// If launching a job, control the spec here
	//+optional
	Job Job `json:"job"`

	// Release of Hyperqueue to installed (if hq binary not found in PATH)
	// +kubebuilder:default="0.15.0"
	// +default="0.15.0"
	// +optional
	HyperqueueVersion string `json:"hyperqueueVersion,omitempty"`

	// Size of the hyperqueue (1 server + (N-1) nodes)
	Size int32 `json:"size"`

	// Interactive mode keeps the cluster running
	// +optional
	Interactive bool `json:"interactive"`

	// Time limit for the job
	// Approximately one year. This cannot be zero or job won't start
	// +kubebuilder:default=31500000
	// +default=31500000
	// +optional
	DeadlineSeconds int64 `json:"deadlineSeconds,omitempty"`

	// Resources include limits and requests
	// +optional
	Resources Resource `json:"resources"`
}

type Job struct {

	// Nodes for the job (defaults to 0 for 1)
	// +optional
	Nodes int64 `json:"nodes"`

	// Name for the job
	// +optional
	Name string `json:"name"`

	// Name for the log file
	// +optional
	Log string `json:"log"`
}

// Node corresponds to a pod (server or worker)
type Node struct {

	// Image to use for hyperqueue
	// +kubebuilder:default="ubuntu"
	// +default="ubuntu"
	// +optional
	Image string `json:"image"`

	// Port for hyperqueue to use.
	// Since we have a headless service, this
	// is not represented in the operator, just
	// in starting the server or a worker
	// +optional
	Port int32 `json:"port"`

	// Resources include limits and requests
	// +optional
	Resources Resources `json:"resources"`

	// PullSecret for the node, if needed
	// +optional
	PullSecret string `json:"pullSecret"`

	// Command will be honored by a server node
	// +optional
	Command string `json:"command,omitempty"`

	// Commands to run around different parts of the hyperqueu setup
	// +optional
	Commands Commands `json:"commands,omitempty"`

	// Working directory
	// +optional
	WorkingDir string `json:"workingDir,omitempty"`

	// PullAlways will always pull the container
	// +optional
	PullAlways bool `json:"pullAlways"`

	// Ports to be exposed to other containers in the cluster
	// We take a single list of integers and map to the same
	// +optional
	// +listType=atomic
	Ports []int32 `json:"ports"`

	// Key/value pairs for the environment
	// +optional
	Environment map[string]string `json:"environment"`
}

// ContainerResources include limits and requests
type Commands struct {

	// Init runs before anything in both scripts
	// +optional
	Init string `json:"init,omitempty"`
}

// ContainerResources include limits and requests
type Resources struct {

	// +optional
	Limits Resource `json:"limits"`

	// +optional
	Requests Resource `json:"requests"`
}

type Resource map[string]intstr.IntOrString

// Validate the Hyperqueue
func (hq *Hyperqueue) Validate() bool {

	// These are fairly arbitrary
	if hq.Spec.Server.Port == 0 {
		hq.Spec.Server.Port = 6789
	}
	if hq.Spec.Worker.Port == 0 {
		hq.Spec.Worker.Port = 1234
	}
	// TODO cannot compare to empty structure later!
	if hq.Spec.Worker.Image == "" {
		hq.Spec.Worker.Image = hq.Spec.Server.Image
	}
	if hq.Spec.ServiceName == "" {
		hq.Spec.ServiceName = "hq-service"
	}
	if hq.Spec.Job.Name == "" {
		hq.Spec.Job.Name = "hq-job"
	}
	return true
}

// WorkerNodes returns the number of worker nodes
// At this point we've already validated size is >= 1
func (hq *Hyperqueue) WorkerNodes() int32 {
	return hq.Spec.Size - 1
}

// WorkerNode returns the worker node (if defined) or falls back to the server
func (hq *Hyperqueue) WorkerNode() Node {

	// If we don't have a worker spec, copy the parent
	workerNode := hq.Spec.Worker
	if reflect.DeepEqual(workerNode, Node{}) {
		workerNode = hq.Spec.Server
	}
	return workerNode
}

// HyperqueueStatus defines the observed state of Hyperqueue
type HyperqueueStatus struct{}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Hyperqueue is the Schema for the hyperqueues API
type Hyperqueue struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   HyperqueueSpec   `json:"spec,omitempty"`
	Status HyperqueueStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// HyperqueueList contains a list of Hyperqueue
type HyperqueueList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Hyperqueue `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Hyperqueue{}, &HyperqueueList{})
}
