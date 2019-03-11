/*
Copyright 2018 Endocode AG.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"flag"
	"os"
	"reflect"
	"time"

	"github.com/golang/glog"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	// Uncomment the following line to load the gcp plugin (only required to authenticate against GKE clusters).
	// _ "k8s.io/client-go/plugin/pkg/client/auth/gcp"

	clientset "github.com/endocode/kelefstis/pkg/client/clientset/versioned"
	informers "github.com/endocode/kelefstis/pkg/client/informers/externalversions"
)

var (
	masterURL  string
	kubeconfig string
)

func main() {
	flag.Set("logtostderr", "true")
	flag.Set("v", "1")
	flag.Set("kubeconfig", os.Getenv("HOME")+"/.kube/config")
	flag.Parse()

	cfg, err := clientcmd.BuildConfigFromFlags(masterURL, kubeconfig)
	if err != nil {
		glog.Fatalf("Error building kubeconfig: %s", err.Error())
	}

	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		glog.Fatalf("Error building kubernetes clientset: %s", err.Error())
	}

	exampleClient, err := clientset.NewForConfig(cfg)
	if err != nil {
		glog.Fatalf("Error building example clientset: %s", err.Error())
	}

	/* 	kif := kubeinformers.NewSharedInformerFactory(kubeClient, time.Second*30)
	   	rpi := reflect.ValueOf(kif.Core().V1()).MethodByName("Pods").Call([]reflect.Value{})
			 pi := rpi[0].Interface().(coreinformers.PodInformer) */
	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(kubeClient, time.Second*30)

	informerFactory := func(name string) interface{} {
		rpi := reflect.ValueOf(kubeInformerFactory.Core().V1()).MethodByName(name).Call([]reflect.Value{})
		return rpi[0].Interface()
	}
	exampleInformerFactory := informers.NewSharedInformerFactory(exampleClient, time.Second*30)
	controller := NewController(kubeClient, exampleClient,
		informerFactory,
		exampleInformerFactory.Samplecontroller().V1alpha1().RuleCheckers())

	go kubeInformerFactory.Start(controller.stopCh)
	go exampleInformerFactory.Start(controller.stopCh)

	if err = controller.Run(2, controller.stopCh); err != nil {
		glog.Fatalf("Error running controller: %s", err.Error())
	}
}

func init() {
	flag.StringVar(&kubeconfig, "kubeconfig", "", "Path to a kubeconfig. Only required if out-of-cluster.")
	flag.StringVar(&masterURL, "master", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
}
