# Kelefstis

Kelefstis "the boatsman" is an example howto use a Kubernetes controller to
do compliance checks from inside Kubernetes.

It has been derived from the original [sample-controller](https://github.com/kubernetes/sample-controller)

## Installation notes

For a fast deployment, use the `install.sh` script in the `deployment` folder

```bash
cd deployment
./install.sh
```

### Dependencies

There are no specical vendor depencies, however, implicitely it relies on a `k8s.io/client-go`
version 1.15, this version uses `HEAD kubernetes-1.15.1-beta.0`

### Building the Image

The image is build without dependencies, just go into the `image` folder
and run `k7s-image-from-scratch.sh`

```bash
cd image
k7s-image-from-scratch.sh
```

### Compiling the code

`go build` in this directory works out of the box,
go-get or vendor this package as `github.com/endocode/kelefstis`.
The installation with `go get github.com/endocode/kelefstis` will need some time,
because a lot of Kubernetes code and other libs are involved but should work out of the box.

## Purpose

This repository implements a simple controller for watching Pod resources
by a `RuleChecker`. Look into the `      (artifacts/examples) directory for definitions and basic examples.

This particular example demonstrates how to perform one basic operation:

* check all the pod images against a matching regular expression.
* anything else is not tested, yet
* test have been performed against [minikube v1.2.0](https://kubernetes.io/docs/setup/minikube/)

## Removed

It makes no longer use of the generators in [k8s.io/code-generator](https://github.com/kubernetes/code-generator). The typical `./hack/update-codegen.sh` ... script are not used.

The [goju library](https://github.com/endocode/goju) is not longer used to implement checks
of JSON or YAML definitions by these rules.

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
    - kind: "Pod"
      apiVersion: "v1"
      metadata:
        namespace:
          eq: "lirumlarum"
      spec:
        containers:
        - image:
            matches: "\
              (k8s.gcr.io|\
              gcr.io|\
              quay.io/kubernetes-ingress-controller|\
              quay.io/endocode|\
              quay.io/coreos|\
              docker.io/istio|\
              docker.io/prom)\
              "
            securityContext:
              privileged:
                equals: false
                must: true
        initContainers:
        - image:
            matches: "^(k8s.gcr.io|gcr.io|quay.io/kubernetes-ingress-controller|quay.io/endocode|quay.io/coreos|docker.io/istio|docker.io/prom)"
            securityContext:
              privileged:
                equals: false
                must: true
    - kind: "Cluster"
      apiVersion: "some/v1beta"
      spec:
        min: 3
        max: 10
    - kind: "Node"
      apiVersion: "v1"
      status:
        allocatable:
          cpu:
            min: "2"
          pods:
            max: "200"

```

## Details

The sample controller uses the `unstructured` library of [client-go library](https://github.com/kubernetes/client-go/) extensively.
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
0808 09:46:43.461372   11022 main.go:197] Creating watch
I0808 09:46:43.462124   11022 main.go:206] Creating channel
I0808 09:46:43.478340   11022 main.go:60] add: kelefstis.endocode.com/v1alpha1:RuleChecker/default/rules
I0808 09:46:43.478406   11022 main.go:60] add: kelefstis.endocode.com/v1alpha1:RuleChecker/default/rules
I0808 09:46:43.478414   11022 main.go:60] add: kelefstis.endocode.com/v1alpha1:RuleChecker/default/rules
I0808 09:46:43.478422   11022 main.go:156] adding rule kelefstis.endocode.com/v1alpha1:RuleChecker/default/rules
I08
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
I0807 20:10:57.596563   13136 main.go:197] Creating watch
I0807 20:10:57.597407   13136 main.go:206] Creating channel
I0807 20:10:57.651252   13136 main.go:60] add: kelefstis.endocode.com/v1alpha1:RuleChecker/default/rules
I0807 20:10:57.651301   13136 main.go:60] add: kelefstis.endocode.com/v1alpha1:RuleChecker/default/rules
I0807 20:10:57.651321   13136 main.go:60] add: kelefstis.endocode.com/v1alpha1:RuleChecker/default/rules
I0807 20:10:57.651339   13136 main.go:156] adding rule kelefstis.endocode.com/v1alpha1:RuleChecker/default/rules
I0807 20:10:57.701636   13136 treecheck.go:45] add: v1:Pod/kube-system/kube-controller-manager-kind-control-plane2
I0807 20:10:57.702097   13136 result.go:200] result Matches((k8s.gcr.io|gcr.io|quay.io/kubernetes-ingress-controller|quay.io/endocode|quay.io/coreos|docker.io/istio|docker.io/prom),k8s.gcr.io/kube-controller-manager:v1.15.0) = (true, ok)
I0807 20:10:57.702182   13136 result.go:70] checking object v1:Pod/kube-system/kube-controller-manager-kind-control-plane2 by rule kelefstis.endocode.com/v1alpha1:RuleChecker/default/rules: 2 true, 0 false
I0807 20:10:57.702209   13136 treecheck.go:45] add: v1:Pod/kube-system/kube-proxy-4bqhc
I0807 20:10:57.702386   13136 result.go:200] result Matches((k8s.gcr.io|gcr.io|quay.io/kubernetes-ingress-controller|quay.io/endocode|quay.io/coreos|docker.io/istio|docker.io/prom),k8s.gcr.io/kube-proxy:v1.15.0) = (true, ok)
I0807 20:10:57.702436   13136 result.go:70] checking object v1:Pod/kube-system/kube-proxy-4bqhc by rule kelefstis.endocode.com/v1alpha1:RuleChecker/default/rules: 2 true, 0 false
I0807 20:10:57.702458   13136 treecheck.go:45] add: v1:Pod/kube-system/kindnet-8nzd7
I0807 20:10:57.702627   13136 result.go:200] result Matches((k8s.gcr.io|gcr.io|quay.io/kubernetes-ingress-controller|quay.io/endocode|quay.io/coreos|docker.io/istio|docker.io/prom),kindest/kindnetd:0.5.0) = (false, ok)
I0807 20:10:57.702673   13136 result.go:70] checking object v1:Pod/kube-system/kindnet-8nzd7 by rule kelefstis.endocode.com/v1alpha1:RuleChecker/default/rules: 0 true, 2 false
I0807 20:10:57.702694   13136 treecheck.go:45] add: v1:Pod/kube-system/kube-proxy-xzmqq
I0807 20:10:57.702869   13136 result.go:200] result Matches((k8s.gcr.io|gcr.io|quay.io/kubernetes-ingress-controller|quay.io/endocode|quay.io/coreos|docker.io/istio|docker.io/prom),k8s.gcr.io/kube-proxy:v1.15.0) = (true, ok)
I0807 20:10:57.702903   13136 result.go:70] checking object v1:Pod/kube-system/kube-proxy-xzmqq by rule kelefstis.endocode.com/v1alpha1:RuleChecker/default/rules: 2 true, 0 false
I0807 20:10:57.702924   13136 treecheck.go:45] add: v1:Pod/kube-system/kube-scheduler-kind-control-plane3
I0807 20:10:57.703096   13136 result.go:200] result Matches((k8s.gcr.io|gcr.io|quay.io/kubernetes-ingress-controller|quay.io/endocode|quay.io/coreos|docker.io/istio|docker.io/prom),k8s.gcr.io/kube-scheduler:v1.15.0) = (true, ok)
I0807 20:10:57.703149   13136 result.go:70] checking object v1:Pod/kube-system/kube-scheduler-kind-control-plane3 by rule kelefstis.endocode.com/v1alpha1:RuleChecker/default/rules: 2 true, 0 false
I0807 20:10:57.703170   13136 treecheck.go:45] add: v1:Pod/kube-system/kube-scheduler-kind-control-plane
I0807 20:10:57.703371   13136 result.go:200] result Matches((k8s.gcr.io|gcr.io|quay.io/kubernetes-ingress-controller|quay.io/endocode|quay.io/coreos|docker.io/istio|docker.io/prom),k8s.gcr.io/kube-scheduler:v1.15.0) = (true, ok)
I0807 20:10:57.703414   13136 result.go:70] checking object v1:Pod/kube-system/kube-scheduler-kind-control-plane by rule kelefstis.endocode.com/v1alpha1:RuleChecker/default/rules: 2 true, 0 false
I0807 20:10:57.703434   13136 treecheck.go:45] add: v1:Pod/kube-system/etcd-kind-control-plane2
I0807 20:10:57.703615   13136 result.go:200] result Matches((k8s.gcr.io|gcr.io|quay.io/kubernetes-ingress-controller|quay.io/endocode|quay.io/coreos|docker.io/istio|docker.io/prom),k8s.gcr.io/etcd:3.3.10) = (true, ok)
I0807 20:10:57.703657   13136 result.go:70] checking object v1:Pod/kube-system/etcd-kind-control-plane2 by rule kelefstis.endocode.com/v1alpha1:RuleChecker/default/rules: 2 true, 0 false
I0807 20:10:57.703682   13136 treecheck.go:45] add: v1:Pod/kube-system/etcd-kind-control-plane
I0807 20:10:57.703849   13136 result.go:200] result Matches((k8s.gcr.io|gcr.io|quay.io/kubernetes-ingress-controller|quay.io/endocode|quay.io/coreos|docker.io/istio|docker.io/prom),k8s.gcr.io/etcd:3.3.10) = (true, ok)
I0807 20:10:57.703901   13136 result.go:70] checking object v1:Pod/kube-system/etcd-kind-control-plane by rule kelefstis.endocode.com/v1alpha1:RuleChecker/default/rules: 2 true, 0 false
I0807 20:10:57.703929   13136 treecheck.go:45] add: v1:Pod/kube-system/etcd-kind-control-plane3
I0807 20:10:57.704101   13136 result.go:200] result Matches((k8s.gcr.io|gcr.io|quay.io/kubernetes-ingress-controller|quay.io/endocode|quay.io/coreos|docker.io/istio|docker.io/prom),k8s.gcr.io/etcd:3.3.10) = (true, ok)
I0807 20:10:57.704140   13136 result.go:70] checking object v1:Pod/kube-system/etcd-kind-control-plane3 by rule kelefstis.endocode.com/v1alpha1:RuleChecker/default/rules: 2 true, 0 false
I0807 20:10:57.704161   13136 treecheck.go:45] add: v1:Pod/kube-system/kindnet-lvz8c
I0807 20:10:57.704345   13136 result.go:200] result Matches((k8s.gcr.io|gcr.io|quay.io/kubernetes-ingress-controller|quay.io/endocode|quay.io/coreos|docker.io/istio|docker.io/prom),kindest/kindnetd:0.5.0) = (false, ok)
I0807 20:10:57.704392   13136 result.go:70] checking object v1:Pod/kube-system/kindnet-lvz8c by rule kelefstis.endocode.com/v1alpha1:RuleChecker/default/rules: 0 true, 2 false
I0807 20:10:57.704414   13136 treecheck.go:45] add: v1:Pod/kube-system/kube-apiserver-kind-control-plane3
I0807 20:10:57.704589   13136 result.go:200] result Matches((k8s.gcr.io|gcr.io|quay.io/kubernetes-ingress-controller|quay.io/endocode|quay.io/coreos|docker.io/istio|docker.io/prom),k8s.gcr.io/kube-apiserver:v1.15.0) = (true, ok)
I0807 20:10:57.704639   13136 result.go:70] checking object v1:Pod/kube-system/kube-apiserver-kind-control-plane3 by rule kelefstis.endocode.com/v1alpha1:RuleChecker/default/rules: 2 true, 0 false
I0807 20:10:57.704662   13136 treecheck.go:45] add: v1:Pod/kube-system/kube-proxy-tthj2
I0807 20:10:57.704836   13136 result.go:200] result Matches((k8s.gcr.io|gcr.io|quay.io/kubernetes-ingress-controller|quay.io/endocode|quay.io/coreos|docker.io/istio|docker.io/prom),k8s.gcr.io/kube-proxy:v1.15.0) = (true, ok)
I0807 20:10:57.704873   13136 result.go:70] checking object v1:Pod/kube-system/kube-proxy-tthj2 by rule kelefstis.endocode.com/v1alpha1:RuleChecker/default/rules: 2 true, 0 false
I0807 20:10:57.704894   13136 treecheck.go:45] add: v1:Pod/kube-system/kindnet-vvwpg
I0807 20:10:57.705084   13136 result.go:200] result Matches((k8s.gcr.io|gcr.io|quay.io/kubernetes-ingress-controller|quay.io/endocode|quay.io/coreos|docker.io/istio|docker.io/prom),kindest/kindnetd:0.5.0) = (false, ok)
I0807 20:10:57.705125   13136 result.go:70] checking object v1:Pod/kube-system/kindnet-vvwpg by rule kelefstis.endocode.com/v1alpha1:RuleChecker/default/rules: 0 true, 2 false
I0807 20:10:57.705161   13136 treecheck.go:45] add: v1:Pod/kube-system/kube-controller-manager-kind-control-plane
I0807 20:10:57.705336   13136 result.go:200] result Matches((k8s.gcr.io|gcr.io|quay.io/kubernetes-ingress-controller|quay.io/endocode|quay.io/coreos|docker.io/istio|docker.io/prom),k8s.gcr.io/kube-controller-manager:v1.15.0) = (true, ok)
I0807 20:10:57.705378   13136 result.go:70] checking object v1:Pod/kube-system/kube-controller-manager-kind-control-plane by rule kelefstis.endocode.com/v1alpha1:RuleChecker/default/rules: 2 true, 0 false
I0807 20:10:57.705399   13136 treecheck.go:45] add: v1:Pod/kube-system/kube-apiserver-kind-control-plane2
I0807 20:10:57.705593   13136 result.go:200] result Matches((k8s.gcr.io|gcr.io|quay.io/kubernetes-ingress-controller|quay.io/endocode|quay.io/coreos|docker.io/istio|docker.io/prom),k8s.gcr.io/kube-apiserver:v1.15.0) = (true, ok)
I0807 20:10:57.705635   13136 result.go:70] checking object v1:Pod/kube-system/kube-apiserver-kind-control-plane2 by rule kelefstis.endocode.com/v1alpha1:RuleChecker/default/rules: 2 true, 0 false
I0807 20:10:57.705656   13136 treecheck.go:45] add: v1:Pod/kube-system/kube-apiserver-kind-control-plane
I0807 20:10:57.705827   13136 result.go:200] result Matches((k8s.gcr.io|gcr.io|quay.io/kubernetes-ingress-controller|quay.io/endocode|quay.io/coreos|docker.io/istio|docker.io/prom),k8s.gcr.io/kube-apiserver:v1.15.0) = (true, ok)
I0807 20:10:57.706240   13136 result.go:70] checking object v1:Pod/kube-system/kube-apiserver-kind-control-plane by rule kelefstis.endocode.com/v1alpha1:RuleChecker/default/rules: 2 true, 0 false
I0807 20:10:57.706285   13136 treecheck.go:45] add: v1:Pod/kube-system/kube-proxy-j7f75
I0807 20:10:57.706480   13136 result.go:200] result Matches((k8s.gcr.io|gcr.io|quay.io/kubernetes-ingress-controller|quay.io/endocode|quay.io/coreos|docker.io/istio|docker.io/prom),k8s.gcr.io/kube-proxy:v1.15.0) = (true, ok)
I0807 20:10:57.706518   13136 result.go:70] checking object v1:Pod/kube-system/kube-proxy-j7f75 by rule kelefstis.endocode.com/v1alpha1:RuleChecker/default/rules: 2 true, 0 false
I0807 20:10:57.706539   13136 treecheck.go:45] add: v1:Pod/kube-system/kube-controller-manager-kind-control-plane3
I0807 20:10:57.706710   13136 result.go:200] result Matches((k8s.gcr.io|gcr.io|quay.io/kubernetes-ingress-controller|quay.io/endocode|quay.io/coreos|docker.io/istio|docker.io/prom),k8s.gcr.io/kube-controller-manager:v1.15.0) = (true, ok)
I0807 20:10:57.706754   13136 result.go:70] checking object v1:Pod/kube-system/kube-controller-manager-kind-control-plane3 by rule kelefstis.endocode.com/v1alpha1:RuleChecker/default/rules: 2 true, 0 false
I0807 20:10:57.706774   13136 treecheck.go:45] add: v1:Pod/kube-system/kindnet-vgx8s
I0807 20:10:57.706945   13136 result.go:200] result Matches((k8s.gcr.io|gcr.io|quay.io/kubernetes-ingress-controller|quay.io/endocode|quay.io/coreos|docker.io/istio|docker.io/prom),kindest/kindnetd:0.5.0) = (false, ok)
I0807 20:10:57.706997   13136 result.go:70] checking object v1:Pod/kube-system/kindnet-vgx8s by rule kelefstis.endocode.com/v1alpha1:RuleChecker/default/rules: 0 true, 2 false
I0807 20:10:57.707022   13136 treecheck.go:45] add: v1:Pod/kube-system/coredns-5c98db65d4-b4jq2
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
