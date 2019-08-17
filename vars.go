package main

import (
	"flag"
	"os"

	"github.com/golang/glog"
	"github.com/kubernetes/client-go/kubernetes/scheme"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
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

func initCfg() (err error) {
	g := schema.GroupVersion{Group: "", Version: "v1"}
	cfg.GroupVersion = &g

	cfg.APIPath = "/api"
	if cfg.UserAgent == "" {
		cfg.UserAgent = "k7s/" + restclient.DefaultKubernetesUserAgent()
	}
	/*
		codec, ok := api.Codecs.SerializerForFileExtension("json")
		if !ok {
			return fmt.Errorf("unable to find serializer for JSON")
		}
		cfg.Codec = codec
	*/
	cfg.NegotiatedSerializer = serializer.DirectCodecFactory{CodecFactory: scheme.Codecs}

	if cfg.QPS == 0 {
		cfg.QPS = 5
	}
	if cfg.Burst == 0 {
		cfg.Burst = 10
	}
	return
}

func initFlags() {
	flag.Set("logtostderr", "true")
	flag.Set("v", "1")

	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig == "" {
		kubeconfig = os.Getenv("HOME") + "/.kube/config"
	}
	flag.Set("kubeconfig", kubeconfig)
	var err error
	cfg, err = clientcmd.BuildConfigFromFlags(masterURL, kubeconfig)
	if err != nil {
		glog.Fatalf("Error building kubeconfig: %s", err.Error())
	}
	flag.Parse()
	initCfg()
}
