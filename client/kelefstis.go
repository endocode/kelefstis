package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

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

func display(s interface{}, t string) {

	reflectType := reflect.TypeOf(s).Elem()
	reflectValue := reflect.ValueOf(s).Elem()

	fmt.Printf("%s#ReflectType=%s NumField=%d\n", t, reflectType, reflectType.NumField())

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
			display(&valueValue, "\t\t")
		default:
			fmt.Printf("Default: %s%s : %s(%s)\n", t, typeName, valueValue, valueType)
		}

	}
}

func printValue(buf io.Writer, prefix string, path string, template string, v reflect.Value) {
	fmt.Printf("%s%s ", prefix, path)
	// Drill down through pointers and interfaces to get a value we can print.
	for v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
		if v.Kind() == reflect.Ptr {
			// Check for recursive data
			if !v.CanInterface() {
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
			indexedTemplate := fmt.Sprintf("index (%s %d)", template, i)
			printValue(buf, prefix+"\t", fmt.Sprintf("%s[%d]", path, i), indexedTemplate, v.Index(i))
		}
	case reflect.Struct:
		t := v.Type()
		// use type to get number and names of fields
		fmt.Printf("(Struct with %d fields)\n", t.NumField())

		for i := 0; i < t.NumField(); i++ {
			if !t.Field(i).Anonymous {
				fmt.Printf("%s%s: ", prefix, t.Field(i).Name)
				fieldName := "." + t.Field(i).Name
				printValue(buf, prefix+"\t", path+fieldName, template+fieldName, v.Field(i))
			}
			if v.Type() == reflect.TypeOf(Time{}) {
				fmt.Printf("%s%s: %s\n", prefix, t.Field(i).Name, v.Interface())
			}
		}
	case reflect.Invalid:
		fmt.Printf("Invalid!\n")
	case reflect.String:
		//fmt.Fprintf(buf, "%s{{$.MatchString \"heapster\" (index (index ($.Pods \"\").Items 0).Spec.Containers 0).Image}}\n", path)
		//{{ (%s", prefix, path)
		//fmt.Fprintf(buf, ") 0}}")
		fmt.Printf("%s String: %q\n", prefix, v)
		fmt.Printf("%s Template: (%s %s)\n", prefix, template, v.String())
	default:
		{
			if v.CanInterface() {
				fmt.Printf("Default Interface: %v\n", v.Interface())
			} else {
				fmt.Printf("Default: %s\n", v.Type())
			}
		}
	}
}

//ListCRD is to check by a raw template. To be removed
func ListCRD(clientset *kubernetes.Clientset, group string, version string, crd string, resource string) error {
	rules := clientset.
		CoreV1().
		RESTClient().
		Get().
		AbsPath("apis", group, version, crd, resource).
		Do()

	var rchck RuleChecker

	raw, err := rules.Raw()
	if err != nil {
		return err
	}
	var prettyJSON bytes.Buffer
	err = json.Indent(&prettyJSON, raw, "", "\t")
	//	fmt.Printf("\n-------------\n%s\n----------\n", prettyJSON.String())

	json.Unmarshal(raw, &rchck)
	//	display(&rchck,"")
	if err != nil {
		return err
	}
	bufWriter := bytes.NewBufferString("")
	printValue(bufWriter, "", "", "", reflect.ValueOf(rchck))
	checkTemplate := string(bufWriter.Bytes())

	//fmt.Printf(">>>%s<<<\n", checkTemplate)
	split := strings.Split(checkTemplate, "\n")

	chk := Check{Check: true, Clientset: clientset.CoreV1()}

	for _, t := range split {
		fmt.Printf("%s\n", t)
		tmpl, err := template.New("testgen").Parse(t)
		if err != nil {
			panic(err)
		}
		err = tmpl.Execute(os.Stdout, &chk)
	}
	//display(rchck, "")
	/*
			m, err := json.Marshal(rchck)

			fmt.Printf("%24s",m)
			fmt.Printf("\n\n%s\n\n", prettyJSON)

			if err!=nil {
				panic(err.Error())
			}

		y, _ := yaml.JSONToYAML(raw)
		fmt.Printf("\n\n%s", y)
	*/
	return nil
}
