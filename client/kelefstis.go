package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/docopt/docopt-go"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"reflect"
)

//Set fetches the ClientSet and other objects from the arguments and config
func Set(argv []string) (*kubernetes.Clientset, interface{}, interface{}, interface{}, bool, error) {
	usage := `kelefstis.

Usage:
	kelefstis -t <check> [--kubeconfig <config>] [ -d ]
	kelefstis <rules> [-k <kind>] [--kubeconfig <config>] [ -d ]
	kelefstis [ -h | --help ]
Options:
	-h --help             Show this screen.
	check                 Template with the checks to run
	rules                 Name of the rules
	-k kind               Kind of the rules, defaults to rulechecker
	-d 	                  debug
	--kubeconfig <config>
`
	var ok bool
	help := (len(os.Args) == 1)
	if argv == nil && len(os.Args) > 1 {
		argv = os.Args[1:]
	}
	arguments, err := docopt.Parse(usage, argv, false, "kelefstis 0.1", true)
	//	fmt.Fprintf(os.Stderr, "%s\n", arguments)

	if !help {
		help, ok = arguments["-h"].(bool)
	}

	if !help {
		help, ok = arguments["--help"].(bool)
	}

	if help || err != nil {
		fmt.Print(usage)
		os.Exit(0)
	}

	var debug bool
	_, debug = arguments["-d"].(string)

	kubeconfig, ok := arguments["--kubeconfig"].(string)
	if !ok {
		kubeconfig = filepath.Join(os.Getenv("HOME"), ".kube", "config")
	}
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, nil, nil, nil, debug, errors.New("no kubeconfig found")
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, nil, nil, nil, debug, errors.New("kubeconfig invalid")
	}
	check, ok := arguments["<check>"].(string)
	if ok {
		checktemplate, err := ioutil.ReadFile(check)
		return clientset, string(checktemplate), nil, nil, debug, err

	}
	if arguments["<rules>"] != nil {

		kind, ok := arguments["-k"].(string)
		if !ok {
			kind = "rulechecker"
		}

		return clientset, nil, arguments["<rules>"], kind, debug, nil

	}
	return nil, nil, nil, nil, debug, errors.New("this should not happen")
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
	if err != nil {
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

func listResource(clientset *kubernetes.Clientset) {
	raw, err := clientset.CoreV1().
		RESTClient().Get().
		Resource("").DoRaw()

	if err != nil {
		panic(err.Error())
	}
	var prettyJSON bytes.Buffer
	err = json.Indent(&prettyJSON, raw, "", "\t")
	//	fmt.Printf("--------> %-24s\n\n", prettyJSON.String())
	/*if len(nodes.Items) > 0 {
		fmt.Printf("There are %d nodes in the cluster\n", len(nodes.Items))
		for _, node := range nodes.Items {
			fmt.Printf("  Node %-36s\n", node.Name)
		}

	}
	*/
}

//ListCRD is to check by a raw template. To be removed
func ListCRD(clientset *kubernetes.Clientset, group string, version string, crd string, resource string) error {
	rules := clientset.
		CoreV1().
		RESTClient().
		Get().
		AbsPath("apis", group, version, crd, resource).
		Do()

	raw, err := rules.Raw()
	if err != nil {
		return err
	}
	var prettyJSON bytes.Buffer
	err = json.Indent(&prettyJSON, raw, "", "\t")
	//	fmt.Printf("\n-------------\n%s\n----------\n", prettyJSON.String())
	rchck := RuleChecker{}

	json.Unmarshal(raw, &rchck)
	//	display(&rchck,"")
	if err != nil {
		return err
	}
	bufWriter := bytes.NewBufferString("NumberOfPods={{.NumberOfPods \"\"}}\n")

	chk := Check{Check: true, Clientset: clientset.CoreV1()}

	for _, r := range rchck.Spec.Rules {
		chk.PrintTemplate(bufWriter, "", nil, reflect.ValueOf(r))
	}

	checkTemplate := string(bufWriter.Bytes())

	fmt.Printf("%s\n", checkTemplate)

	tmpl, err := template.New("testgen").Parse(checkTemplate)
	if err != nil {
		panic(err)
	}
	err = tmpl.Execute(os.Stdout, &chk)

	return err
}
