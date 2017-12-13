package main

import (
	"fmt"
	"encoding/json"
	"github.com/docopt/docopt-go"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"path/filepath"
	"github.com/endocode/kelefstis/check"
	"text/template"
	"io/ioutil"
	"bytes"
	"github.com/ghodss/yaml"
	"time"
)

// simple k8s client that lists all available pods
// it gets config from $HOME/.kube/config
func main()  {
	usage:=`kelefstis.

Usage:
	kelefstis <check> [--kubeconfig <config>]
	kelefstis [ -h | --help ]

Options:
	-h --help             Show this screen.
		check                 Template with the checks to run
	--kubeconfig <config>
`
	arguments,_ := docopt.Parse(usage,nil,false,"kelefstis 0.1", false)

	if arguments["--help"].(bool)  || arguments["<check>"] == nil {
		fmt.Printf(usage)
		return
	}


	checkfile:=arguments["<check>"].(string)

	kubeconfig, ok := arguments["--kubeconfig"].(string)
	if ! ok {
		kubeconfig = filepath.Join(os.Getenv("HOME"), ".kube", "config")
	}

	checktemplate, err := ioutil.ReadFile(checkfile)
	if err != nil {
		fmt.Printf(usage)
		return
	}

	fmt.Println("Using kubeconfig: ", kubeconfig)
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	chk := check.Check{true, clientset.CoreV1()}
	tmpl, err := template.New("test").Parse(string(checktemplate))
	if err != nil {
		panic(err)
	}
	err = tmpl.Execute(os.Stdout, &chk)
	if err != nil {
		panic(err.Error())
	}
	/*
	listNodes(clientset)
	listPods(clientset)
	listResource(clientset)
	*/
	listCRD(clientset, "stable.example.com", "v1", "rulecheckers","rules")

}
func listPods(clientset *kubernetes.Clientset) (*apiv1.PodList, error) {
	pods, err := clientset.CoreV1().Pods("").List(metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}
	if len(pods.Items) > 0 {
		fmt.Printf("There are %d pods in the cluster\n", len(pods.Items))
		for _, pod := range pods.Items {
			fmt.Printf("  Pod %-36s %-36s -%48s\n", pod.Name, pod.Namespace, pod.Labels)
		}

	}
	return pods, err
}

func listNodes(clientset *kubernetes.Clientset) (*apiv1.NodeList, error) {
	nodes, err := clientset.CoreV1().Nodes().List(metav1.ListOptions{})
	if err!=nil {
		panic(err.Error())
	}
	if len(nodes.Items) > 0 {
		fmt.Printf("There are %d nodes in the cluster\n", len(nodes.Items))
		for _, node := range nodes.Items {
			fmt.Printf("  Node %-36s\n", node.Name)
		}

	}
	return nodes, err
}

func listResource(clientset *kubernetes.Clientset)  {
	raw ,err := clientset.CoreV1().
		RESTClient().Get().
		Resource("").DoRaw()

	if err!=nil {
		panic(err.Error())
	}
	var prettyJSON bytes.Buffer
	err= json.Indent(&prettyJSON, raw, "", "\t")
	fmt.Printf("--------> %-24s\n\n", prettyJSON)
	/*if len(nodes.Items) > 0 {
		fmt.Printf("There are %d nodes in the cluster\n", len(nodes.Items))
		for _, node := range nodes.Items {
			fmt.Printf("  Node %-36s\n", node.Name)
		}

	}
*/
}
/*

created with https://mholt.github.io/json-to-go/

 */
type RuleChecker struct {
	APIVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
	Metadata   struct {
		ClusterName                string      `json:"clusterName"`
		CreationTimestamp          time.Time   `json:"creationTimestamp"`
		DeletionGracePeriodSeconds interface{} `json:"deletionGracePeriodSeconds"`
		DeletionTimestamp          interface{} `json:"deletionTimestamp"`
		Description                string      `json:"description"`
		Initializers               interface{} `json:"initializers"`
		Name                       string      `json:"name"`
		Namespace                  string      `json:"namespace"`
		ResourceVersion            string      `json:"resourceVersion"`
		SelfLink                   string      `json:"selfLink"`
		UID                        string      `json:"uid"`
	} `json:"metadata"`
	Spec struct {
		Rules []struct {
			Domain   string `json:"domain"`
			MinNodes int    `json:"minNodes,omitempty"`
		} `json:"rules"`
	} `json:"spec"`
}

func listCRD(clientset *kubernetes.Clientset,group string, version string, crd string, resource string) {
	rules := clientset.CoreV1().
		RESTClient().
			Get().
			AbsPath("apis",group,version,crd,resource).
			Do()

	var rchck RuleChecker

	fmt.Printf("%s", &rchck)

	raw, err := rules.Raw()



	var prettyJSON bytes.Buffer
	err= json.Indent(&prettyJSON, raw, "", "\t")
	json.Unmarshal(raw,rchck)
	m,_ := json.Marshal(rchck)
	fmt.Printf("%24s",m)
	//	fmt.Printf("\n\n%s\n\n", prettyJSON)

	if err!=nil {
		panic(err.Error())
	}
	y, _:=yaml.JSONToYAML(raw)
	fmt.Printf("\n\n%s",y)

}