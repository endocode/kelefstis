package client

import (
	"fmt"
	"io"
	"log"
	"os"
	"reflect"
	"regexp"
	"strings"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	clientv1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

//A Check is the result of all the checkcs
//Clientset the K8S interface to use
type Check struct {
	Check     bool
	Clientset clientv1.CoreV1Interface
	nodeList  *apiv1.NodeList
	podList   *apiv1.PodList
}

func init() {
	log.SetOutput(os.Stderr)
}

//NotZero checks if the argument is non-zero
func (chk *Check) NotZero(i int) *Check {
	chk.Check = chk.Check && i != 0
	return chk
}

//GE checks if the first argument is bigger than the second
func (chk *Check) GE(n int, m int) *Check {
	chk.Check = chk.Check && n > m
	return chk
}

//MatchString checks, if the string s matches the regular expression r
func (chk *Check) MatchString(r string, s string) *Check {
	match, _ := regexp.MatchString(r, s)
	log.Printf("%s %s %t", r, s, match)

	chk.Check = chk.Check && match
	return chk
}

//ToString uses fmt.Sprintf to convert anything to a string
func (*Check) ToString(i interface{}) string {
	return fmt.Sprintf("%s", i)
}

//Eq checks, if the string s is equal to r
func (chk *Check) Eq(r string, s string) *Check {

	eq := (r == s)
	log.Printf("\nEq: %q %q %t\n", r, s, eq)

	chk.Check = chk.Check && eq
	return chk
}

//Nodes returns the Nodes().List()  forwarded from the Clientset
func (chk *Check) Nodes() ([]apiv1.Node, error) {
	if chk.nodeList == nil {
		nodeList, err := chk.Clientset.Nodes().List(metav1.ListOptions{})
		if err == nil {
			chk.nodeList = nodeList
		}
		return nodeList.Items, err
	}
	return chk.nodeList.Items, nil
}

//NumberOfNodes  returns the number of nodes in the cluster
func (chk *Check) NumberOfNodes() (int, error) {
	list, err := chk.Nodes()

	return len(list), err
}

/*NumberOfPods returns the number of pods in the namespace
'' means the default namespace
*/
func (chk *Check) NumberOfPods(namespace string) (int, error) {
	list, err := chk.Pods(namespace)

	return len(list), err
}

/*Pods returns the Pods().List()  forwarded from the Clientset in the namespace
'' means the default namespace
*/
func (chk *Check) Pods(namespace string) ([]apiv1.Pod, error) {
	if chk.podList == nil {
		podList, err := chk.Clientset.Pods(namespace).List(metav1.ListOptions{})
		if err == nil {
			chk.podList = podList
		}
		return podList.Items, err
	}
	return chk.podList.Items, nil
}

//HasMethod checks if the method has been implemented
func (chk *Check) HasMethod(method string) bool {
	st := reflect.TypeOf(chk)
	_, ok := st.MethodByName(method)
	return ok
}

// PrintTemplate creates a template
func (chk *Check) PrintTemplate(buf io.Writer, prefix string, path []string, v reflect.Value) {
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
			//			fmt.Printf("%s%d: ", prefix, i)
			index := append([]string{"index"}, path...)
			chk.PrintTemplate(buf, prefix+"\t", append(index, fmt.Sprintf("%d", i)), v.Index(i))
		}
	case reflect.Struct:
		t := v.Type()
		// use type to get number and names of fields
		fmt.Printf("(Struct with %d fields)\n", t.NumField())
		rangeFlag := false
		for i := 0; i < t.NumField(); i++ {
			if !t.Field(i).Anonymous {
				if t.Field(i).Name == "Range" {
					sf, found := t.FieldByName("Namespace")
					nameSpace := ""
					if found {
						nameSpace = "\"" + v.Field(sf.Index[0]).String() + "\""
					}
					fmt.Fprintf(buf, "{{ range .%s %s}}\n", strings.Join(path, "."), nameSpace)
					//	fmt.Fprintf(buf, "%s", "{ {printf \"%-24s\" .Image} }")
					rangeFlag = true
					path = nil
				} else {
					fmt.Printf("%s%s: ", prefix, t.Field(i).Name)
					fieldName := t.Field(i).Name
					chk.PrintTemplate(buf, prefix+"\t", append(path, fieldName), v.Field(i))
				}
			}
			if v.Type() == reflect.TypeOf(Time{}) {
				fmt.Printf("%s%s: %s\n", prefix, t.Field(i).Name, v.Interface())
			}
		}
		if rangeFlag {
			fmt.Fprintf(buf, "{{end}}\n")
		}
	case reflect.Invalid:
		fmt.Printf("Invalid!\n")
	case reflect.String:
		//fmt.Fprintf(buf, "%s{{$.MatchString \"heapster\" (index (index ($.Pods \"\").Items 0).Spec.Containers 0).Image}}\n", path)
		//{{ (%s", prefix, path)
		//fmt.Fprintf(buf, ") 0}}")
		fmt.Printf("%s String: %q\n", prefix, v)
		pathString := strings.Join(path[:len(path)-1], ".")
		fs := path[len(path)-1]
		fmt.Fprintf(buf, "(%s %q .%s).Check=", fs, v, pathString)
		if chk.HasMethod(fs) {
			// to convert any argument to a string the printf "%s" is used
			// which returns a string. This seems to be more reliable than
			// s, ok := i.(string) conversion. Weird, but works
			fmt.Fprintf(buf, "{{($.%s %q (printf \"%%s\" .%s)).Check}}\n ", fs, v, pathString)
		} else {
			fmt.Fprintf(buf, "method %q does not exist\n", fs)
		}
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
