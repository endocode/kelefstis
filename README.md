# kelefstis
Extensible Kubernetes compliance check tool

This is an example how to connect to a K8S master from Go using the client api.

A template is used to create a basic report checking some compliance rules.

The name [Kelefstis is Greek](https://en.bab.la/dictionary/english-greek/boatswain) for [Boatswain](http://work.chron.com/duties-boatswain-20927.html), who has the [maintenance role on a ship](http://www.thepirateking.com/historical/ship_roles.htm)

To be tested, there must be a running Kubernetes cluster configured in `.kube/config`

Please extend!

Tests will follow! I promise!

Install it with all deps by `go get -u github.com/endocode/kelefstis`, this will take some time fetching all dependencies.

To run the tests, you need a working Kubernetes cluster, however [minikube](https://github.com/kubernetes/minikube) also works.

As a first step, you need to create a `CustomerResourceDefinition` for the RuleChecker and the rule defined in the yaml files

```

kubectl create -f rulecheckersr-rsd.yaml

kubectl create -f rule.yaml	

```

Now the tests in `templ.yaml` should work

```
kelefstis templ.yaml
```

`kelefstis_test` creates and removes all the objects in the kubernetes test cluster now.
