package main

import (
	"fmt"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	clientv1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"path/filepath"
	"regexp"
	"text/template"
)

type check struct {
	Check     bool
	clientset clientv1.CoreV1Interface
}

func (this *check) NotZero(i int) *check {
	this.Check = this.Check && i != 0
	return this
}

func (this *check) GE(n int, m int) *check {
	this.Check = this.Check && n > m
	return this
}

func (this *check) MatchString(r string, s string) *check {
	match, _ := regexp.MatchString(r, s)
	this.Check = this.Check && match
	return this
}

func (this *check) Nodes() (*apiv1.NodeList, error) {
	return this.clientset.Nodes().List(metav1.ListOptions{})
}

func (this *check) Pods(namespace string) (*apiv1.PodList, error) {
	return this.clientset.Pods(namespace).List(metav1.ListOptions{})
}

// simple k8s client that lists all available pods
// it gets config from $HOME/.kube/config
func main() {
	checktemplate := `we found {{(len .Nodes.Items)}} nodes={{(.NotZero (len .Nodes.Items) ).Check}}
we found {{len (.Pods "").Items}} pods
{{range (.Pods "").Items}}
Pod {{printf "%-36s" .GetName}} {{printf "%-24s" .Status.Phase}}{{.ClusterName}}{{range .Spec.Containers}}
    Container {{printf "%-24s" .Name}} {{printf "%-24s" .Image}} ({{printf "%t" ($.MatchString "^gcr.io/google[-_]containers" .Image).Check}}) {{.SecurityContext}}{{end}}{{end}}

Result: Compliance {{.Check}}
`
	kubeconfig := filepath.Join(os.Getenv("HOME"), ".kube", "config")
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

	nodes, err := clientset.CoreV1().Nodes().List(metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}
	chk := check{true, clientset.CoreV1()}
	if len(nodes.Items) > 0 {
		fmt.Printf("There are %d nodes in the cluster\n", len(nodes.Items))
		for _, node := range nodes.Items {
			fmt.Printf("  Node %-36s\n", node.Name)
		}

	}
	tmpl, err := template.New("test").Parse(checktemplate)
	if err != nil {
		panic(err)
	}
	err = tmpl.Execute(os.Stdout, &chk)
	if err != nil {
		panic(err.Error())
	}
}
