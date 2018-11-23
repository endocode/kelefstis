package main

import (
	"github.com/endocode/goju"
	"github.com/golang/glog"
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
	rulemap, err := CRD(clientset, "apis", "stable.example.com", "v1", "rulecheckers", "", "rules")
	if err != nil {
		panic(err)
	}
	for k, v := range rulemap {

		glog.V(2).Infof("\nusing rule for %s:\n%s\n", k, v)

		items, err := CRD(clientset, "api", "", "v1", k, "", k)
		if err != nil {
			glog.V(1).Infof("error %s", err)
		}
		var treecheck = goju.TreeCheck{Check: &goju.Check{}}

		treecheck.Traverse(items, rulemap)
		glog.V(0).Infof("tests error: %d", treecheck.ErrorHistory.Len())
		glog.V(0).Infof("tests true : %d", treecheck.TrueCounter)
		glog.V(0).Infof("tests false: %d", treecheck.FalseCounter)
	}
}
