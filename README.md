# Kelefstis

Kelefstis "the boatsman" is an example howto use a Kubernetes controller to
do compliance checks from inside Kubernetes.

It has been derived from the original [sample-controller](https://github.com/kubernetes/sample-controller)

## Installation notes

Go-get or vendor this package as `github.com/endocode/kelefstis`.
The installation with `go get github.com/endocode/kelefstis` will need some time,
because a lot of Kubernetes code and other libs are involved but should work out of the box.

## Purpose

This repository implements a simple controller for watching Pod resources
by a `RuleChecker`. Look into the `      (artifacts/examples) directory for definitions and basic examples.

This particular example demonstrates how to perform one basic operation:

* check all the pod images against a matching regular expression.
* anything else is not tested, yet
* test have been performed against [minikube v1.10](https://kubernetes.io/docs/setup/minikube/)

It makes use of the generators in [k8s.io/code-generator](https://github.com/kubernetes/code-generator)
to generate a typed client, informers, listers and deep-copy functions. You can
do this yourself using the `./hack/update-codegen.sh` script.

The `update-codegen` script will automatically generate the following files &
directories:

* `pkg/apis/kelefstis/v1alpha1/zz_generated.deepcopy.go`
* `pkg/client/`

Changes should not be made to these files manually, and when creating your own
controller based off of this implementation you should not copy these files and
instead run the `update-codegen` script to generate your own.

This is an example of how to build use a controller and do simple checks from the inside.
The principal structure of a `RuleChecker` show, that the definition follows the definition
of the pods.

The final leave of a definition like `..pod.spec.containers.image`
is augmented by a check, here the `matches` declaration.

```yaml
apiVersion: kelefstis.endocode.com/v1alpha1
kind: RuleChecker
metadata:
  name: rules
  description: "my cluster, my rules"
spec:
  rules:
    * pods:
          range: "all"
          namespace:
            eq: "lirumlarum"
          spec:
            containers:
              range: "all"
              image:
                matches: "^(k8s.gcr.io|gcr.io|quay.io/kubernetes-ingress-controller/nginx-ingress-controller)"
    * cluster:
        min: 3
        max: 10
    * nodes:
        memory:
          min: 100GB
```

The [goju library](https://github.com/endocode/goju) is used to implement checks
of JSON or YAML definitions by these rules.

## Details

The sample controller uses [client-go library](https://github.com/kubernetes/client-go/tree/master/tools/cache) extensively.
The details of interaction points of the sample controller with various mechanisms from this library are
explained [here](docs/controller-client-go.md).

## Limitations

Currently, only the match is implemented as a kind of `Hello World` to demonstrate
at least one useful purpose. There might be issues by bad defined yaml files if
numbers are used were strings are expectd.

## Running

**Prerequisite**: Since the sample-controller uses `apps/v1` deployments,
the Kubernetes cluster version should be greater than 1.9.

```sh
# assumes you have a working kubeconfig, not required if operating in-cluster
$ go build
```

Create the `RuleChecker` CRD

```sh
$ kubectl create -f artifacts/examples/rulecheckers-crd.yaml
customresourcedefinition.apiextensions.k8s.io "rulecheckers.kelefstis.endocode.com" created
```

Start `kelefstis` logging to stderr with your config.
A Useful loglevel is `-v 2`, with higher values you get more verbose output.

```sh
$ ./kelefstis -alsologtostderr -kubeconfig ~/.kube/config -v 2
I1124 19:25:19.539835   29935 controller.go:149] Setting up event handlers
I1124 19:25:19.540030   29935 controller.go:234] Starting controller
I1124 19:25:19.540040   29935 controller.go:237] Waiting for informer caches to sync
```

Create some rules from a different terminal

```sh
$ kubectl apply -f artifacts/examples/rules.yaml
rulechecker.kelefstis.endocode.com "rules" created
```

This matches all image names `k8s.gcr.io`,
with version minikube v1.10 of 13 of 16 containers are flagged as matching. The `rule.alt.yaml` file
matches `gcr.io` only, accepting 15 containers. Finally, `rule.all` matches all 16 containers with the
best regexp `^(k8s.gcr.io|gcr.io|quay.io/kubernetes-ingress-controller/nginx-ingress-controller)`

```sh
...
I1124 18:28:35.043910    5628 controller.go:171] RuleChecker changed:
namespace:
  eq: default
range: all
spec:
  containers:
    image:
      matches: gcr.io
    range: all
I1124 18:28:35.049328    5628 goju.go:79] #1: ..spec.containers[0].image.Matches("gcr.io","gcr.io/k8s-minikube/storage-provisioner:v1.8.1"): true
I1124 18:28:35.057331    5628 goju.go:79] #2: ..spec.containers[0].image.Matches("gcr.io","k8s.gcr.io/heapster-amd64:v1.5.3"): true
I1124 18:28:35.059445    5628 controller.go:203] Pod nginx-ingress-controller-8566746984-jbxnc changed to:
I1124 18:28:35.059696    5628 controller.go:203] Pod influxdb-grafana-5l2fc changed to:
I1124 18:28:35.059823    5628 controller.go:203] Pod kube-scheduler-minikube changed to:
I1124 18:28:35.059952    5628 controller.go:203] Pod kube-dns-86f4d74b45-77m5n changed to:
I1124 18:28:35.060028    5628 controller.go:203] Pod default-http-backend-544569b6d7-5c8qn changed to:
I1124 18:28:35.060040    5628 controller.go:203] Pod etcd-minikube changed to:
I1124 18:28:35.060051    5628 controller.go:203] Pod kube-addon-manager-minikube changed to:
I1124 18:28:35.060060    5628 controller.go:203] Pod kube-proxy-prhqv changed to:
I1124 18:28:35.060069    5628 controller.go:203] Pod kube-controller-manager-minikube changed to:
I1124 18:28:35.060080    5628 controller.go:203] Pod kube-apiserver-minikube changed to:
I1124 18:28:35.060091    5628 controller.go:203] Pod kubernetes-dashboard-6f4cfc5d87-4h4kh changed to:
I1124 18:28:35.060106    5628 controller.go:203] Pod storage-provisioner changed to:
I1124 18:28:35.060120    5628 controller.go:203] Pod heapster-7dcb9 changed to:
I1124 18:28:35.062425    5628 goju.go:79] #3: ..spec.containers[0].image.Matches("gcr.io","k8s.gcr.io/kube-addon-manager:v8.6"): true
I1124 18:28:35.064582    5628 goju.go:79] #4: ..spec.containers[0].image.Matches("gcr.io","k8s.gcr.io/kube-proxy-amd64:v1.10.0"): true
I1124 18:28:35.066631    5628 goju.go:79] #5: ..spec.containers[0].image.Matches("gcr.io","k8s.gcr.io/kube-controller-manager-amd64:v1.10.0"): true
I1124 18:28:35.068721    5628 goju.go:79] #6: ..spec.containers[0].image.Matches("gcr.io","k8s.gcr.io/kube-apiserver-amd64:v1.10.0"): true
I1124 18:28:35.070504    5628 goju.go:79] #7: ..spec.containers[0].image.Matches("gcr.io","k8s.gcr.io/kubernetes-dashboard-amd64:v1.10.0"): true
I1124 18:28:35.073835    5628 goju.go:79] #8: ..spec.containers[0].image.Matches("gcr.io","gcr.io/google_containers/defaultbackend:1.4"): true
I1124 18:28:35.075446    5628 goju.go:79] #9: ..spec.containers[0].image.Matches("gcr.io","k8s.gcr.io/etcd-amd64:3.1.12"): true
I1124 18:28:35.077445    5628 goju.go:79] #10: ..spec.containers[0].image.Matches("gcr.io","quay.io/kubernetes-ingress-controller/nginx-ingress-controller:0.19.0"): false
I1124 18:28:35.079813    5628 goju.go:79] #11: ..spec.containers[0].image.Matches("gcr.io","k8s.gcr.io/heapster-influxdb-amd64:v1.3.3"): true
I1124 18:28:35.079863    5628 goju.go:79] #12: ..spec.containers[1].image.Matches("gcr.io","k8s.gcr.io/heapster-grafana-amd64:v4.4.3"): true
I1124 18:28:35.081979    5628 goju.go:79] #13: ..spec.containers[0].image.Matches("gcr.io","k8s.gcr.io/kube-scheduler-amd64:v1.10.0"): true
I1124 18:28:35.087990    5628 goju.go:79] #14: ..spec.containers[0].image.Matches("gcr.io","k8s.gcr.io/k8s-dns-kube-dns-amd64:1.14.8"): true
I1124 18:28:35.088039    5628 goju.go:79] #15: ..spec.containers[1].image.Matches("gcr.io","k8s.gcr.io/k8s-dns-dnsmasq-nanny-amd64:1.14.8"): true
I1124 18:28:35.088079    5628 goju.go:79] #16: ..spec.containers[2].image.Matches("gcr.io","k8s.gcr.io/k8s-dns-sidecar-amd64:1.14.8"): true
I1124 18:28:35.088100    5628 controller.go:182] Errors       : 0
I1124 18:28:35.088107    5628 controller.go:183] Checks   true: 15
I1124 18:28:35.088113    5628 controller.go:184] Checks  false: 1
...
```

## Use case

In a typical project only certain registries are allowed, however,
there might be examples where only selected images from
other registries are trusted. You can define a regexp explicitely
expressing trust in these images. In the example above, the build in images
of minikube are trusted, which are hosted mostly in `k8s.gcr.io` and
`gcr.io`. The one exception `quay.io/kubernetes-ingress-controller/nginx-ingress-controller` must
be explicitely named.

## Check for privileged containers

Running [Istio](https://istio.io/) in the standard configuration it is obvious, 
that by adding a privileged container to the pod everybody can circumvent the entire security 
of the [Envoy](https://www.envoyproxy.io/) sidecar. This is a well documented [issue](https://github.com/istio/old_issues_repo/issues/172).

Therefore, Istio can only be run securely with the [Istio CNI Plugin](https://github.com/istio/cni#istio-cni-plugin), which is not the default.

To check for running privileged containers, the rule
```yaml
...
spec:
  rules:
    - pods:
          ...
          spec:
            containers:
              ...
              securityContext:
                privileged:
                  equals: false
...
```
should be used.

## Cleanup

You can clean up the created CustomResourceDefinition with:

```sh
$ kubectl delete crd rulecheckers.kelefstis.endocode.com
customresourcedefinition.apiextensions.k8s.io "rulecheckers.kelefstis.endocode.com" deleted
```

## Outlook

The plan is to create more and more sophisticated checks.

* Basic checks like
  * more Kubernetes entities
  * sizes, limits
  * namespace specific policies
  * node checks
  * ...
* Sophisticated checks
  * RBAC rules
  * NetworkPolicies
  * connectivity checks
  * ...
* Reactive Mode
  * remove objects not following the rules
  * or fix them
  * use exec to do checks inside a container
  * ...