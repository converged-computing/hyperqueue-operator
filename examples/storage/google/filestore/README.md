# Filestore (NFS)

This tutorial will walk through creating a more persistent Hyperqueue cluster on Google Cloud
using Filestore. We will be following the guidance [here](https://cloud.google.com/filestore/docs/csi-driver).
First, make sure you have the Filestore and GKE (Google Kubernetes Engine) APIs enabled,
and the other introductory steps at that link. If you've used Google Cloud for Kubernetes and
Filestore before, you should largely be good to go.

## Create the Cluster

First, create your cluster with the Filestore CSI Driver enabled:

```bash
GOOGLE_PROJECT=myproject
```
```bash
$ gcloud container clusters create test-cluster --project $GOOGLE_PROJECT \
    --zone us-central1-a --machine-type n1-standard-2 \
    --addons=GcpFilestoreCsiDriver \
    --num-nodes=4 --enable-network-policy --tags=test-cluster --enable-intra-node-visibility
```

Create the Flux Operator namespace:

```bash
$ kubectl create namespace hyperqueue-operator
```

## Install the Hyperqueue Operator

Let's next install the operator. We first need the jobs API:

```bash
VERSION=v0.2.0
kubectl apply --server-side -f https://github.com/kubernetes-sigs/jobset/releases/download/$VERSION/manifests.yaml
```

Then we can grab the latest from the repository:

```bash
$ kubectl apply -f https://raw.githubusercontent.com/converged-computing/hyperqueue-operator/main/examples/dist/hyperqueue-operator.yaml
```

Next, we want to create our persistent volume claim. This will use the
storage drivers installed already to our cluster (via the creation command).

```yaml
kind: PersistentVolumeClaim
apiVersion: v1
metadata:
  name: data
  namespace: hyperqueue-operator
spec:
  accessModes:
    - ReadWriteMany
  resources:
    requests:
      storage: 1Ti
  storageClassName: standard-rwx
```

Note that we've selected the storage class "standard-rwx" from [this list](https://cloud.google.com/filestore/docs/csi-driver#storage-class).
You can also see the `storageclass` available for your cluster via this command:

```bash
$ kubectl get storageclass
```

<details>

<summary>storageclass Available on our Filestore Cluster</summary>

```console
NAME                        PROVISIONER                    RECLAIMPOLICY   VOLUMEBINDINGMODE      ALLOWVOLUMEEXPANSION   AGE
enterprise-multishare-rwx   filestore.csi.storage.gke.io   Delete          WaitForFirstConsumer   true                   16m
enterprise-rwx              filestore.csi.storage.gke.io   Delete          WaitForFirstConsumer   true                   16m
premium-rwo                 pd.csi.storage.gke.io          Delete          WaitForFirstConsumer   true                   15m
premium-rwx                 filestore.csi.storage.gke.io   Delete          WaitForFirstConsumer   true                   16m
standard                    kubernetes.io/gce-pd           Delete          Immediate              true                   15m
standard-rwo (default)      pd.csi.storage.gke.io          Delete          WaitForFirstConsumer   true                   15m
standard-rwx                filestore.csi.storage.gke.io   Delete          WaitForFirstConsumer   true                   16m
```

</details>

A Filestore storage class will *not* be the default (see above output) so this step is important to take! We 
are going to create a persistent volume claim that says "Please use the `standard-rwx` storageclass from
Filestore to be available as a persistent volume claim - and I want all of it - the entire 1TB!

```bash
$ kubectl apply -f pvc.yaml
```
And check on the status:

```bash
$ kubectl get -n hyperqueue-operator pvc
NAME   STATUS    VOLUME   CAPACITY   ACCESS MODES   STORAGECLASS   AGE
data   Pending                                      standard-rwx   6s
```

It will be pending under we make a request to use it! Let's do that by creating the Hyperqueue cluster:

```bash
$ kubectl apply -f hyperqueue.yaml
```
It will take a hot minute to pull the container for the access pod, and then I found (maybe because of Filestore?)
the second level of containers (server and worker set) didn't start pulling until almost 3 minutes.
You can watch progress as follows:

```bash
kubectl get -n hyperqueue-operator pods
```
```console
$ kubectl get -n hyperqueue-operator pods
NAME                                 READY   STATUS      RESTARTS   AGE
hyperqueue-sample-access             0/1     Completed   0          5m31s
hyperqueue-sample-server-0-0-wf77c   1/1     Running     0          4m2s
hyperqueue-sample-worker-0-0-862mn   1/1     Running     0          4m2s
hyperqueue-sample-worker-0-1-vpwbf   1/1     Running     0          4m2s
```

## Test your Storage

Let's shell into the server pod and test our Filesystem!

```bash
$ kubectl exec -it -n hyperqueue-operator hyperqueue-sample-server-0-0-wf77c bash
$ ls /workflow/
lost+found

$ touch /workflow/test.txt
$ ls /workflow/
lost+found  test.txt
```
Is it there? How big is it?

```bash
$ df -a | grep /workflow
10.75.133.26:/vol1 1055763456       0 1002059776   0% /workflow
```
Now exit and shell into another pod... a worker one:

```bash
$ kubectl exec -it -n hyperqueue-operator hyperqueue-sample-worker-0-0-g29vs bash
# ls /workflow/
lost+found  test.txt
```
We have a dinosaur fart! I repeat - we have a dinosaur fart!! 🦖🌬️

## Run LAMMPS

Let's now run LAMMPS (this is from the server node)

```bash
hq submit --nodes 2 --log --name test-lammps /workflow/test.out mpirun -np 3 --map-by socket lmp -v x 2 -v y 2 -v z 2 -in in.reaxc.hns -nocite
Job submitted successfully, job ID: 2
```

You should be able to see the job with `hq job list` and then view the log:

```
$ hq job list
+----+-------------+---------+-------+
| ID | Name        | State   | Tasks |
+----+-------------+---------+-------+
|  2 | test-lammps | RUNNING | 1     |
+----+-------------+---------+-------+
There are 2 jobs in total. Use `--all` to display all jobs.

$ cat /workflow/test.out 
HQ:log LAMMPS (29 Sep 2021 - Update 2)
ROMP_NUM_THREADS environment is not set. Defaulting to 1 thread. (src/comm.cpp:98)
(  using 1 OpenMP thread(s) per MPI task
Reading data file ...
�  triclinic box = (0.0000000 0.0000000 0.0000000) to (22.326000 11.141200 13.778966) with tilt (0.0000000 -5.0260300 0.0000000)
!  3 by 1 by 1 MPI processor grid
  reading atoms ...

  304 atoms
  reading velocities ...
  304 velocities
   read_data CPU = 0.095 seconds
Replicating atoms ...
�  triclinic box = (0.0000000 0.0000000 0.0000000) to (44.652000 22.282400 27.557932) with tilt (0.0000000 -10.052060 0.0000000)
!  3 by 1 by 1 MPI processor grid
R  bounding box image = (0 -1 -1) to (0 1 1)
  bounding box extra memory = 0.03 MB
?  average # of replicas added to proc = 6.00 out of 8 (75.00%)
  2432 atoms
   replicate CPU = 0.179 seconds
?Neighbor list info ...
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
A  Unit style    : real
  Current step  : 0
  Time step     : 0.1
LPer MPI rank memory allocation (min/avg/max) = 117.7 | 120.3 | 122.0 Mbytes
-Step Temp PotEng Press E_vdwl E_coul Volume 
X       0          300   -113.27833    437.52123   -111.57687   -1.7014647    27418.867 
X      10    299.38517   -113.27631     1439.271   -111.57492   -1.7013813    27418.867 
X      20    300.27107   -113.27884    3764.3835   -111.57762   -1.7012246    27418.867 
X      30    302.21062   -113.28428    7007.6598   -111.58335   -1.7009363    27418.867 
X      40    303.52263   -113.28799    9844.8242   -111.58747   -1.7005186    27418.867 
X      50    301.87058   -113.28324    9663.0414   -111.58318   -1.7000524    27418.867 
X      60    296.67807   -113.26777    7273.8182   -111.56815   -1.6996137    27418.867 
X      70    292.19997   -113.25435    5533.5842   -111.55514   -1.6992157    27418.867 
X      80    293.58675   -113.25831    5993.3772   -111.55946   -1.6988534    27418.867 
X      90    300.62633   -113.27925    7202.8789   -111.58069   -1.6985591    27418.867 
X     100    305.38274   -113.29357    10085.735   -111.59518   -1.6983875    27418.867 
?Loop time of 111.351 on 3 procs for 100 steps with 2432 atoms

@Performance: 0.008 ns/day, 3093.093 hours/ns, 0.898 timesteps/s
259.7% CPU use with 3 MPI tasks x 1 OpenMP threads
�
MPI task timing breakdown:
Section |  min time  |  avg time  |  max time  |%varavg| %total
---------------------------------------------------------------
?Pair    | 18.915     | 21.171     | 23.674     |  42.4 | 19.01
?Neigh   | 0.49588    | 0.6748     | 0.84523    |  17.4 |  0.61
?Comm    | 4.2145     | 6.7168     | 8.9697     |  75.2 |  6.03
?Output  | 0.56624    | 0.57997    | 0.58891    |   1.3 |  0.52
?Modify  | 82.029     | 82.201     | 82.373     |   1.5 | 73.82
?Other   |            | 0.008134   |            |       |  0.01

Nlocal:        810.667 ave         835 max         770 min
Histogram: 1 0 0 0 0 0 0 0 1 1
Nghost:        6564.67 ave        6648 max        6431 min
Histogram: 1 0 0 0 0 0 0 0 1 1
Neighs:        302035.0 ave      309580 max      289379 min
Histogram: 1 0 0 0 0 0 0 0 1 1

pTotal # of neighbors = 906104
Ave neighs/atom = 372.57566
Neighbor list builds = 5
Dangerous builds not checked
Total wall time: 0:01:55
```

And that is it! We will want to test this at a larger scale against the Flux Operator.

## Clean Up

Don't forget to clean up! Delete the MiniCluster and PVC first:

```bash
$ kubectl delete -f examples/storage/google/filestore/hyperqueue.yaml
$ kubectl delete -f examples/storage/google/filestore/pvc.yaml
```

And then delete your Kubernetes cluster. Technically  you could probably just do this, but we might as well be proper!

```bash
$ gcloud container clusters delete --zone us-central1-a test-cluster
```

You will need to also delete the PVC in the Google Cloud storage console.