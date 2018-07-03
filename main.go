package main

import (
	"errors"

	"github.com/endocode/goju"
	"github.com/golang/glog"
	"k8s.io/client-go/kubernetes"
)

func toStringMap(i interface{}) (map[string]interface{}, bool) {
	m, ok := i.(map[string]interface{})
	return m, ok
}
func toArray(i interface{}) ([]interface{}, bool) {
	m, ok := i.([]interface{})
	return m, ok
}

// simple k8s client that lists all available pods
// it gets config from $HOME/.kube/config
func main() {
	initFlags()
	rulemap, err := rules(clientset, "apis", "stable.example.com", "v1", "rulecheckers", "")
	if err != nil {
		panic(err)
	}
	for k, v := range rulemap {

		glog.V(2).Infof("\nusing rule for %s:\n%s\n", k, v)

		_, err := rules(clientset, "api", "", "v1", k, "")
		if err != nil {
			glog.V(1).Infof("error %s", err)
		}
		var treecheck = goju.TreeCheck{}

		treecheck.Traverse(nil, rulemap)
	}
}

func items(clientset *kubernetes.Clientset, path string, group string,
	version string, crd string, resource string) ([]interface{}, error) {

	tree, err := ListCRD(clientset, path, group, version, crd, resource)
	if err != nil {
		return nil, err
	}
	glog.V(1).Infof("\nTree:\n%s\n", tree)

	t, ok := toStringMap(tree)
	if !ok {
		return nil, errors.New("could not convert tree to stringmap")
	}

	items, ok := toArray(t["items"])
	if !ok {
		return nil, errors.New("could not extract items from map")
	}

	return items, nil
}

func rules(clientset *kubernetes.Clientset, path string, group string,
	version string, crd string, resource string) (map[string]interface{}, error) {

	items, err := items(clientset, path, group, version, crd, resource)
	if err != nil {
		return nil, err
	}

	r, ok := toStringMap(items[0])
	if !ok {
		return nil, errors.New("could not extract items[0] from map")
	}

	spec, ok := toStringMap(r["spec"])
	if !ok {
		return nil, errors.New("could not extract spec from map")
	}
	glog.V(2).Infof("\nSpec:\n%s\n", spec)

	rules, ok := toArray(spec["rules"])
	if !ok {
		return nil, errors.New("could not extract rules from spec")
	}
	glog.V(2).Infof("\nRules:\n%s\n", rules[0])

	rules0, ok := toStringMap(rules[0])
	if !ok {
		return nil, errors.New("could not extract rules[0]")
	}
	return rules0, nil
}
