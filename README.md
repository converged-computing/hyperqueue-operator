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
or development version (this is what I did):

```bash
$ kubectl apply --server-side -k github.com/kubernetes-sigs/jobset/config/default?ref=main

# This is right before upgrade to v1alpha2, or June 2nd when I was testing!
# This is also a strategy for deploying a test version
git clone https://github.com/kubernetes-sigs/jobset /tmp/jobset
cd /tmp/jobset
git checkout 93bd85c76fc8afa79b4b5c6d1df9075c99c9f22d
IMAGE_TAG=vanessa/jobset:test make image-build
IMAGE_TAG=vanessa/jobset:test make image-push
IMAGE_TAG=vanessa/jobset:test make deploy
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

Look at the logs to see the worker/server starting:

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

And given that we submit a job with `--wait` and `--log`, our main server will submit the job,
write to a specific output file, and then we can cat it to the terminal. E.g.,:

```bash
$ kubectl logs -n hyperqueue-operator hyperqueue-sample-server-0-0-r8vbl -f
```
```console
Hello, I am a server with hyperqueue-sample-server-0-0
Found extra command mpirun -np 2 --map-by socket lmp -v x 2 -v y 2 -v z 2 -in in.reaxc.hns -nocite
2023-06-17T19:10:39Z INFO No online server found, starting a new server
2023-06-17T19:10:39Z INFO Storing access file as '/root/.hq-server/001/access.json'
+------------------+-------------------------------------------------------------------------------+
| Server directory | /root/.hq-server                                                              |
| Server UID       | A7niLo                                                                        |
| Client host      | hyperqueue-sample-server-0-0.hq-service.hyperqueue-operator.svc.cluster.local |
| Client port      | 6789                                                                          |
| Worker host      | hyperqueue-sample-server-0-0.hq-service.hyperqueue-operator.svc.cluster.local |
| Worker port      | 1234                                                                          |
| Version          | 0.15.0-dev                                                                    |
| Pid              | 18                                                                            |
| Start date       | 2023-06-17 19:10:39 UTC                                                       |
+------------------+-------------------------------------------------------------------------------+
hq submit --wait --name lammps --nodes 2 --log log.out mpirun -np 2 --map-by socket lmp -v x 2 -v y 2 -v z 2 -in in.reaxc.hns -nocite
Job submitted successfully, job ID: 1
2023-06-17T19:10:41Z INFO Worker 1 registered from 10.244.0.94:36124
2023-06-17T19:10:41Z INFO Worker 2 registered from 10.244.0.95:49076
Wait finished in 12s 147ms 683us 588ns: 1 job finished
HQ:logrLAMMPS (29 Sep 2021 - Update 2)
OMP_NUM_THREADS environment is not set. Defaulting to 1 thread. (src/comm.cpp:98)
(  using 1 OpenMP thread(s) per MPI task
Reading data file ...
�  triclinic box = (0.0000000 0.0000000 0.0000000) to (22.326000 11.141200 13.778966) with tilt (0.0000000 -5.0260300 0.0000000)
5  2 by 1 by 1 MPI processor grid
  reading atoms ...
%  304 atoms
  reading velocities ...
1  304 velocities
  read_data CPU = 0.003 seconds
Replicating atoms ...
�  triclinic box = (0.0000000 0.0000000 0.0000000) to (44.652000 22.282400 27.557932) with tilt (0.0000000 -10.052060 0.0000000)
  2 by 1 by 1 MPI processor grid
  bounding box image = (0 -1 -1) to (0 1 1)
  bounding box extra memory = 0.03 MB
?  average # of replicas added to proc = 5.00 out of 8 (62.50%)
-  2432 atoms
  replicate CPU = 0.001 seconds
�Neighbor list info ...
  update every 20 steps, delay 0 steps, check no
  max neighbors/atom: 2000, page size: 100000
  master list distance cutoff = 11
  ghost atom cutoff = 11
  binsize = 5.5, bins = 10 5 6
  2 neighbor lists, perpetual/occasional/extra = 2 0 0
  (1) pair reax/c, perpetual
      attributes: half, newton off, ghost
      pair build: half/bin/newtoff/ghost
      stencil: full/ghost/bin/3d
      bin: standard
  (2) fix qeq/reax, perpetual, copy from (1)
      attributes: half, newton off, ghost
      pair build: copy
      stencil: none
      bin: none
Setting up Verlet run ...
  Unit style    : real
  Current step  : 0
  Time step     : 0.1
yPer MPI rank memory allocation (min/avg/max) = 143.9 | 143.9 | 143.9 Mbytes
Step Temp PotEng Press E_vdwl E_coul Volume 
X       0          300   -113.27833    437.52118   -111.57687   -1.7014647    27418.867 
X      10    299.38517   -113.27631    1439.2824   -111.57492   -1.7013813    27418.867 
X      20    300.27107   -113.27884     3764.342   -111.57762   -1.7012247    27418.867 
X      30    302.21063   -113.28428    7007.6629   -111.58335   -1.7009363    27418.867 
X      40    303.52265   -113.28799    9844.8245   -111.58747   -1.7005186    27418.867 
X      50    301.87059   -113.28324    9663.0973   -111.58318   -1.7000523    27418.867 
X      60    296.67807   -113.26777    7273.8119   -111.56815   -1.6996137    27418.867 
X      70    292.19999   -113.25435    5533.5522   -111.55514   -1.6992158    27418.867 
X      80    293.58677   -113.25831    5993.4438   -111.55946   -1.6988533    27418.867 
X      90    300.62635   -113.27925    7202.8369   -111.58069   -1.6985592    27418.867 
�     100    305.38276   -113.29357    10085.805   -111.59518   -1.6983874    27418.867 
Loop time of 11.6816 on 2 procs for 100 steps with 2432 atoms

Performance: 0.074 ns/day, 324.490 hours/ns, 8.560 timesteps/s
99.9% CPU use with 2 MPI tasks x 1 OpenMP threads

MPI task timing breakdown:
Section |  min time  |  avg time  |  max time  |%varavg| %total
---------------------------------------------------------------
Pair    | 8.1671     | 8.47       | 8.7728     |  10.4 | 72.51
Neigh   | 0.2549     | 0.25591    | 0.25693    |   0.2 |  2.19
Comm    | 0.0087991  | 0.31171    | 0.61462    |  54.3 |  2.67
Output  | 0.0011636  | 0.0012047  | 0.0012459  |   0.1 |  0.01
Modify  | 2.641      | 2.6421     | 2.6432     |   0.1 | 22.62
Other   |            | 0.0007353  |            |       |  0.01

Nlocal:        1216.00 ave        1216 max        1216 min
Histogram: 2 0 0 0 0 0 0 0 0 0
Nghost:        7591.50 ave        7597 max        7586 min
Histogram: 1 0 0 0 0 0 0 0 0 1
Neighs:        432912.0 ave      432942 max      432882 min
Histogram: 1 0 0 0 0 0 0 0 0 1

Total # of neighbors = 865824
Ave neighs/atom = 356.01316
Neighbor list builds = 5
Dangerous builds not checked
Total wall time: 0:00:12
```

Since our job sets interactive: true, this means the cluster stays running after the job is finished,
and we can interactively shell in and submit a job, e.g.,:

```bash
$ kubectl exec -it -n hyperqueue-operator hyperqueue-sample-server-0-0-bbbh2 bash
$ mpirun -np 2 --map-by socket lmp -v x 2 -v y 2 -v z 2 -in in.reaxc.hns -nocite
Job submitted successfully, job ID: 1
```

It should be RUNNING fairly quickly:

```bash
$ hq job list
+----+------+---------+-------+
| ID | Name | State   | Tasks |
+----+------+---------+-------+
|  1 | lmp  | RUNNING | 1     |
+----+------+---------+-------+
```
When it's done it will pop off the queue, and you can add `--all`

```bash
$ hq job list --all
+----+------+----------+-------+
| ID | Name | State    | Tasks |
+----+------+----------+-------+
|  1 | lmp  | FINISHED | 1     |
+----+------+----------+-------+
```

If you don't specify a `--log` file, depending on where you run it, the logs can show up on any worker,
typically in the same working directory in a directory called `log-N` (e.g., log-4). And that's it!
When you are finished:

```bash
$ kind delete cluster
```

### How it works

This project aims to follow the Kubernetes [Operator pattern](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/).

It uses [Controllers](https://kubernetes.io/docs/concepts/architecture/controller/),
which provide a reconcile function responsible for synchronizing resources until the desired state is reached on the cluster.

### Developer

There are a lot of arguments and ways we can customize submit (that likely we want to experiment with):

```
# hq submit --help
Submit a job to HyperQueue

Usage: hq submit [OPTIONS] <COMMANDS>...

Arguments:
  <COMMANDS>...
          Command that should be executed by each task

Options:
      --nodes <NODES>
          Number of nodes; 0
          
          [default: 0]

      --cpus <CPUS>
          Number and placement of CPUs for each job

      --resource <RESOURCE>
          Generic resource request in the form <NAME>=<AMOUNT>

      --time-request <TIME_REQUEST>
          Minimal lifetime of the worker needed to start the job
          
          [default: 0ms]

      --name <NAME>
          Name of the job

      --pin <PIN>
          Pin the job to the cores specified in `--cpus`
          
          [possible values: taskset, omp]

      --cwd <CWD>
          Working directory for the submitted job. The path must be accessible from worker nodes [default: %{SUBMIT_DIR}]

      --stdout <STDOUT>
          Path where the standard output of the job will be stored. The path must be accessible from worker nodes

      --stderr <STDERR>
          Path where the standard error of the job will be stored. The path must be accessible from worker nodes

      --env <ENV>
          Specify additional environment variable for the job. You can pass this flag multiple times to pass multiple variables
          
          `--env=KEY=VAL` - set an environment variable named `KEY` with the value `VAL`

      --each-line <EACH_LINE>
          Create a task array where a task will be created for each line of the given file. The corresponding line will be passed to the task in environment variable `HQ_ENTRY`

      --from-json <FROM_JSON>
          Create a task array where a task will be created for each item of a JSON array stored in the given file. The corresponding item from the array will be passed as a JSON string to the task in environment variable `HQ_ENTRY`

      --array <ARRAY>
          Create a task array where a task will be created for each number in the specified number range. Each task will be passed an environment variable `HQ_TASK_ID`.
          
          `--array=5` - create task array with one job with task ID 5
          
          `--array=3-5` - create task array with three jobs with task IDs 3, 4, 5

      --max-fails <MAX_FAILS>
          Maximum number of permitted task failures. If this limit is reached, the job will fail immediately

      --priority <PRIORITY>
          Priority of each task
          
          [default: 0]

      --time-limit <TIME_LIMIT>
          Time limit per task. E.g. --time-limit=10min

      --log <LOG>
          Stream the output of tasks into this log file

      --task-dir
          Create a temporary directory for task, path is provided in HQ_TASK_DIR The directory is automatically deleted when task is finished

      --crash-limit <CRASH_LIMIT>
          Limits how many times may task be in a running state while worker is lost. If the limit is reached, the task is marked as failed. If the limit is zero, the limit is disabled
          
          [default: 5]

      --wait
          Wait for the job to finish

      --progress
          Interactively observe the progress of the submitted job

      --stdin
          Capture stdin and start the task with the given stdin; the job will be submitted when the stdin is closed

      --directives <DIRECTIVES>
          Select directives parsing mode.
          
          `auto`: Directives will be parsed if the suffix of the first command is ".sh".
           `file`: Directives will be parsed regardless of the first command extension.
           `stdin`: Directives will be parsed from standard input passed to `hq submit` instead from the submitted command.
           `off`: Directives will not be parsed.
          
          
          If enabled, HQ will parse `#HQ` directives from a file located in the first entered command. Parameters following the `#HQ` prefix will be used as parameters for `hq submit`.
          
          Example (script.sh):
           #!/bin/bash
           #HQ --name my-job
           #HQ --cpus=2
           
           program --foo=bar
          
          
          [default: auto]
          [possible values: auto, file, stdin, off]

  -h, --help
          Print help (see a summary with '-h')

GLOBAL OPTIONS:
      --server-dir <SERVER_DIR>
          Path to a directory that stores HyperQueue access files
          
          [env: HQ_SERVER_DIR=]

      --colors <COLORS>
          Console color policy
          
          [default: auto]
          [possible values: auto, always, never]

      --output-mode <OUTPUT_MODE>
          How should the output of the command be formatted
          
          [env: HQ_OUTPUT_MODE=]
          [default: cli]
          [possible values: cli, json, quiet]

      --debug
          Turn on a more detailed log output
          
          [env: HQ_DEBUG=]
```

## License

HPCIC DevTools is distributed under the terms of the MIT license.
All new contributions must be made under this license.

See [LICENSE](https://github.com/converged-computing/cloud-select/blob/main/LICENSE),
[COPYRIGHT](https://github.com/converged-computing/cloud-select/blob/main/COPYRIGHT), and
[NOTICE](https://github.com/converged-computing/cloud-select/blob/main/NOTICE) for details.

SPDX-License-Identifier: (MIT)

LLNL-CODE- 842614