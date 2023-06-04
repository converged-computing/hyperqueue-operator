# hyperqueue-operator

> What happens when I run out of things to do on a Saturday... ohno 

This will be an operator that attempts to use [hyperqueue](https://github.com/It4innovations/hyperqueue) to create a cluster to run tasks.
This isn't tested or working yet (and I'm new to the tool so please don't expect it to work yet). Thank you!

## Development

### Creation

```bash
mkdir hyperqueue-operator
cd hyperqueue-operator/
operator-sdk init --domain flux-framework.org --repo github.com/converged-computing/hyperqueue-operator
operator-sdk create api --version v1alpha1 --kind Hyperqueue --resource --controller
```

## Getting Started

You’ll need a Kubernetes cluster to run against. You can use [KIND](https://sigs.k8s.io/kind) to get a local cluster for testing, or run against a remote cluster.
**Note:** Your controller will automatically use the current context in your kubeconfig file (i.e. whatever cluster `kubectl cluster-info` shows).

### Running on the cluster

Create a cluster with kind:

```bash
$ kind create cluster
```

You'll need to install the jobset API, which eventually will be added to Kubernetes proper (but is not yet!)

```bash
VERSION=v0.1.3
kubectl apply --server-side -f https://github.com/kubernetes-sigs/jobset/releases/download/$VERSION/manifests.yaml
```
or devlopment version:

```bash
$ kubectl apply --server-side -k github.com/kubernetes-sigs/jobset/config/default?ref=main
```

Generate the custom resource definition

```bash
# Build and push the image, and generate the examples/dist/hyperqueue-operator-dev.yaml
$ make test-deploy DEVIMG=<some-registry>/hyperqueue-operator:tag

# As an example
$ make test-deploy DEVIMG=vanessa/hyperqueue-operator:test
```

Make our namespace:

```bash
$ kubectl create namespace hyperqueue-operator
```

Apply the new config!

```bash
$ kubectl apply -f examples/dist/hyperqueue-operator-dev.yaml
```

See logs for the operator

```bash
$ kubectl logs -n hyperqueue-operator-system hyperqueue-operator-controller-manager-6f6945579-9pknp 
```

Create a "hello-world" interactive cluster:

```bash
$ kubectl apply -f examples/tests/hello-world/hyperqueue.yaml 
```

Look at the logs to see the worker/server starting (and hello world)

```console
2023-06-04T06:03:50Z INFO No online server found, starting a new server
2023-06-04T06:03:50Z INFO Saving access file as '/root/.hq-server/001/access.json'
File exists (os error 17)
+------------------+------------------------------------------------+
| Server directory | /root/.hq-server                               |
| Server UID       | L4E27M                                         |
| Host             | hyperqueue-sample-hyperqueue-sample-server-0-0 |
| Pid              | 2710                                           |
| HQ port          | 39795                                          |
| Workers port     | 38937                                          |
| Start date       | 2023-06-04 06:03:50 UTC                        |
| Version          | 0.15.0                                         |
+------------------+------------------------------------------------+
```

Note that we are currently trying to address [this issue](https://github.com/It4innovations/hyperqueue/issues/592) before the worker and server
can properly communicate. The operator cannot assume a shared filesystem. When you are done, cleanup.

```bash
$ kind delete cluster
```

### How it works

This project aims to follow the Kubernetes [Operator pattern](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/).

It uses [Controllers](https://kubernetes.io/docs/concepts/architecture/controller/),
which provide a reconcile function responsible for synchronizing resources until the desired state is reached on the cluster.


### TODO

- Add script logging levels / quiet

## License

HPCIC DevTools is distributed under the terms of the MIT license.
All new contributions must be made under this license.

See [LICENSE](https://github.com/converged-computing/cloud-select/blob/main/LICENSE),
[COPYRIGHT](https://github.com/converged-computing/cloud-select/blob/main/COPYRIGHT), and
[NOTICE](https://github.com/converged-computing/cloud-select/blob/main/NOTICE) for details.

SPDX-License-Identifier: (MIT)

LLNL-CODE- 842614
