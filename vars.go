package main

import (
	"flag"
	"os"

	"github.com/golang/glog"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func init() {
	flag.StringVar(&kubeconfig, "kubeconfig", "", "Path to a kubeconfig. Only required if out-of-cluster.")
	flag.StringVar(&masterURL, "master", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
	//check = TreeCheck{}
	results = NewResultMap()
	ruleMap = make(map[string]*unstructured.Unstructured)
	initFlags()
}

var (
	masterURL  string
	kubeconfig string
	//check      TreeCheck
	results *ResultMap
	ruleMap map[string]*unstructured.Unstructured
	cfg     *restclient.Config
)

func initFlags() {
	flag.Set("logtostderr", "true")
	flag.Set("v", "1")

	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig == "" {
		kubeconfig = os.Getenv("HOME") + "/.kube/config"
	}
	flag.Set("kubeconfig", kubeconfig)
	flag.Parse()
	var err error
	cfg, err = clientcmd.BuildConfigFromFlags(masterURL, kubeconfig)
	if err != nil {
		glog.Fatalf("Error building kubeconfig: %s", err.Error())
	}
}
