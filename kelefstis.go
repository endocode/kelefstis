package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/golang/glog"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// Client returns a kubernetes client from configuration file
func Client(kubeconfig string) (*kubernetes.Clientset, error) {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return clientset, nil
}

var rulechecker, kubeconfig string
var clientset *kubernetes.Clientset

func initFlags() {
	kubeconfig := filepath.Join(os.Getenv("HOME"), ".kube", "config")

	flag.StringVar(&rulechecker, "testrulechecker", "", "check rules")
	flag.StringVar(&kubeconfig, "kubeconfig", kubeconfig, "the place of the kubeconfig")
	flag.Set("logtostderr", "true")
	flag.Set("v", "1")
	flag.Parse()
	var err error
	clientset, err = Client(kubeconfig)

	if err != nil {
		panic(err)
	}
}

func listPods(clientset *kubernetes.Clientset) (*apiv1.PodList, error) {
	pods, err := clientset.CoreV1().Pods("").List(metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}
	if len(pods.Items) > 0 {
		fmt.Printf("There are %d pods in the cluster\n", len(pods.Items))
		for _, pod := range pods.Items {
			fmt.Printf("  Pod %-36s %-36s\n", pod.Name, pod.Namespace)
			for k, v := range pod.Labels {
				fmt.Printf("     %s=%s\n", k, v)
			}
		}

	}
	return pods, err
}

func listResource(clientset *kubernetes.Clientset) {
	raw, err := clientset.CoreV1().
		RESTClient().Get().
		Resource("").DoRaw()

	if err != nil {
		panic(err.Error())
	}

	s, err := prettyJSON(raw)

	fmt.Printf(s)
}
func prettyJSON(raw []byte) (string, error) {
	var buffer bytes.Buffer
	err := json.Indent(&buffer, raw, "", "  ")
	return buffer.String(), err
}

// LogJSON adds structured JSON logging to the Verbose type
func LogJSON(level glog.Level, raw []byte) error {
	json, err := prettyJSON(raw)
	if err != nil {
		glog.V(level).Infof("cannot convert to JSON %s", err)
		return err
	}

	if glog.V(level) {
		lines := strings.Split(json, "\n")
		size := 1 + int(math.Log10(float64(len(lines))))
		format := "%" + strconv.Itoa(size) + "d: %s"
		for i, line := range lines {
			glog.V(level).Infof(format, i, line)
		}
	}
	return nil
}

// ListCRD list a customer resource definition
func ListCRD(clientset *kubernetes.Clientset, path string, group string, version string, crd string, resource string) (interface{}, error) {
	rules := clientset.
		CoreV1().
		RESTClient().
		Get().
		AbsPath(path, group, version, crd, resource).
		Do()

	raw, err := rules.Raw()
	LogJSON(2, raw)
	if err != nil {
		return nil, err
	}
	var tree map[string]interface{}
	json.Unmarshal(raw, &tree)

	return tree, nil
}

// Items returns the *resource* list items
func Items(clientset *kubernetes.Clientset, path string, group string,
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

// CRD goes deeper
func CRD(clientset *kubernetes.Clientset, path string, group string,
	version string, crd string, resource string, key string) (map[string]interface{}, error) {

	items, err := Items(clientset, path, group, version, crd, resource)
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

	rules, ok := toArray(spec[key])
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
