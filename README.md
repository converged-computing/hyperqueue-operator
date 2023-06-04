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
operator-sdk create api --group hyperqueues --version v1alpha1 --kind Hyperqueue --resource --controller
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

Generate the custom resource definition

```bash
# Build and push the image, and generate the examples/dist/hyperqueue-operator-dev.yaml
$ make test-deploy DEVIMG=<some-registry>/hyperqueue-operator:tag
```

Apply the new config!

```bash
$ kubectl apply -f examples/dist/hyperqueue-operator-dev.yaml
```

See logs for the operator

```bash
$ kubectl logs -n hyperqueue-operator-system  hyperqueue-operator-controller-manager-6f6945579-9pknp 
```


When you are done, cleanup.

```bash
$ kind delete cluster
```



### How it works

This project aims to follow the Kubernetes [Operator pattern](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/).

It uses [Controllers](https://kubernetes.io/docs/concepts/architecture/controller/),
which provide a reconcile function responsible for synchronizing resources until the desired state is reached on the cluster.


## License

HPCIC DevTools is distributed under the terms of the MIT license.
All new contributions must be made under this license.

See [LICENSE](https://github.com/converged-computing/cloud-select/blob/main/LICENSE),
[COPYRIGHT](https://github.com/converged-computing/cloud-select/blob/main/COPYRIGHT), and
[NOTICE](https://github.com/converged-computing/cloud-select/blob/main/NOTICE) for details.

SPDX-License-Identifier: (MIT)

LLNL-CODE- 842614
