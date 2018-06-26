package main

import (
	"flag"
	"os"
	"path/filepath"

	"github.com/endocode/goju"
	"github.com/golang/glog"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
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
	var rulechecker string
	kubeconfig := filepath.Join(os.Getenv("HOME"), ".kube", "config")

	flag.StringVar(&rulechecker, "rulechecker", "", "check rules")
	flag.StringVar(&kubeconfig, "kubeconfig", kubeconfig, "the place of the kubeconfig")
	flag.Set("logtostderr", "true")
	flag.Set("v", "1")
	flag.Parse()
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	//	listPods(clientset)
	//	listResource(clientset)
	tree, err := ListCRD(clientset, "stable.example.com", "v1", "rulecheckers", "rules")
	glog.V(1).Infof("\nTree:\n%s\n", tree)

	t, _ := toStringMap(tree)

	spec, _ := toStringMap(t["spec"])
	glog.V(2).Infof("\nSpec:\n%s\n", spec)

	rules, _ := toArray(spec["rules"])
	glog.V(2).Infof("\nRules:\n%s\n", rules[0])

	rules0, _ := toStringMap(rules[0])
	pods, _ := toStringMap(rules0["pods"])
	glog.V(2).Infof("\nPods:\n%s\n", pods)

	var treecheck = goju.TreeCheck{}

	treecheck.Traverse(nil, nil)
}
