package client

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
	"io/ioutil"
	"bytes"
	"github.com/ghodss/yaml"

	"reflect"
)

func ClientSet(argv []string) (*kubernetes.Clientset,string, error){
	usage:=`kelefstis.

Usage:
	kelefstis <check> [--kubeconfig <config>]
	kelefstis [ -h | --help ]

Options:
	-h --help             Show this screen.
		check                 Template with the checks to run
	--kubeconfig <config>
`
	if argv == nil && len(os.Args) > 1 {
		argv = os.Args[1:]
	}
	arguments,_ := docopt.Parse(usage,argv,false,"kelefstis 0.1", false)

	if arguments["--help"].(bool)  || arguments["<check>"] == nil {
		fmt.Printf(usage)
		os.Exit(1)
	}

	checkfile:=arguments["<check>"].(string)
	checktemplate, err := ioutil.ReadFile(checkfile)

	kubeconfig, ok := arguments["--kubeconfig"].(string)
	if ! ok {
		kubeconfig = filepath.Join(os.Getenv("HOME"), ".kube", "config")
	}

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		fmt.Printf("   unusable, exiting")
		os.Exit(3)
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		os.Exit(4)
	}
	return clientset, string( checktemplate), err
}

func ListPods(clientset *kubernetes.Clientset) (*apiv1.PodList, error) {
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

func ListNodes(clientset *kubernetes.Clientset) (*apiv1.NodeList, error) {
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

func ListResource(clientset *kubernetes.Clientset)  {
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

func display(s interface{}, t string) {

	reflectType := reflect.TypeOf(s).Elem()
	reflectValue := reflect.ValueOf(s).Elem()

	fmt.Printf("%s#ReflectType=%s NumField=%d\n",t, reflectType, reflectType.NumField() )

	for i := 0; i < reflectType.NumField(); i++ {
		typeName := reflectType.Field(i).Name

		valueType := reflectValue.Field(i).Type()
		valueValue := reflectValue.Field(i)
		switch reflectValue.Field(i).Kind() {
		case reflect.String, reflect.Int32:
			fmt.Printf("%s%s : %s(%s)\n", t, typeName, valueValue, valueType)
		case reflect.Array:
			fmt.Printf("%sArray: %s : %s(%s)\n", t, typeName, valueValue, valueType)
		case reflect.Struct:
			fmt.Printf("%s%s : %s\n", t, typeName, valueType)
			display(&valueValue,"\t\t")
		default:
			fmt.Printf("Default: %s%s : %s(%s)\n", t, typeName, valueValue, valueType)
		}

	}
}

func printValue(prefix string, v reflect.Value) {

	fmt.Printf("%s: ", v.Type())

	// Drill down through pointers and interfaces to get a value we can print.
	for v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
		if v.Kind() == reflect.Ptr {
			// Check for recursive data
			if ! v.CanInterface() {
				return
			}

		}
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Slice, reflect.Array:
		fmt.Printf("Array %d elements\n", v.Len())
		for i := 0; i < v.Len(); i++ {
			fmt.Printf("%s%d: ", prefix, i)
			printValue(prefix+"\t", v.Index(i))
		}
	case reflect.Struct:
		t := v.Type() // use type to get number and names of fields
		fmt.Printf("%sStruct %d fields\n",prefix, t.NumField())

		for i := 0; i < t.NumField(); i++ {
			if ! t.Field(i).Anonymous {
				fmt.Printf("%s%s: ", prefix, t.Field(i).Name)
				printValue(prefix+"\t", v.Field(i))
			}
			if v.Type() == reflect.TypeOf(Time{}){
				fmt.Printf("%s%s: %s\n", prefix, t.Field(i).Name, v.Interface())
			}
		}
	case reflect.Invalid:
		fmt.Printf("nil\n")
	case reflect.String:
		fmt.Printf("String %v\n", v.Interface())
	default:
		{
			if v.CanInterface() {
				fmt.Printf("default Interface: %v\n", v.Interface())
			} else {
			fmt.Printf("default: %s\n", v.Type())
			}
		}
	}
}


func ListCRD(clientset *kubernetes.Clientset,group string, version string, crd string, resource string) {
	rules := clientset.
		CoreV1().
		RESTClient().
			Get().
			AbsPath("apis",group,version,crd,resource).
			Do()

	var rchck RuleChecker


	raw, err := rules.Raw()



	var prettyJSON bytes.Buffer
	err= json.Indent(&prettyJSON, raw, "", "\t")
	json.Unmarshal(raw,&rchck)
//	display(&rchck,"")
	if err!=nil {
		panic(err.Error())
	}

	printValue("",reflect.ValueOf(rchck))
/*
	m, err := json.Marshal(rchck)

	fmt.Printf("%24s",m)
	fmt.Printf("\n\n%s\n\n", prettyJSON)

	if err!=nil {
		panic(err.Error())
	}
*/
	y, _:=yaml.JSONToYAML(raw)
	fmt.Printf("\n\n%s",y)
}