package client

import (
	"log"
	"os"
	"regexp"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	clientv1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

//A Check is the result of all the checkcs
//Clientset the K8S interface to use
type Check struct {
	Check     bool
	Clientset clientv1.CoreV1Interface
}

func init() {
	log.SetOutput(os.Stderr)
}

//NotZero checks if the argument is non-zero
func (t *Check) NotZero(i int) *Check {
	t.Check = t.Check && i != 0
	return t
}

//GE checks if the first argument is bigger than the second
func (t *Check) GE(n int, m int) *Check {
	t.Check = t.Check && n > m
	return t
}

//MatchString checks, if the string s matches the regular expression r
func (t *Check) MatchString(r string, s string) *Check {
	match, _ := regexp.MatchString(r, s)
	log.Printf("%s %s %t", r, s, match)

	t.Check = t.Check && match
	return t
}

//Nodes returns the Nodes().List()  forwarded from the Clientset
func (t *Check) Nodes() (*apiv1.NodeList, error) {
	return t.Clientset.Nodes().List(metav1.ListOptions{})
}

//NumberOfNodes  returns the number of nodes in the cluster
func (t *Check) NumberOfNodes() (int, error) {
	list, err := t.Nodes()

	return len(list.Items), err
}

/*NumberOfPods returns the number of pods in the namespace
'' means the default namespace
*/
func (t *Check) NumberOfPods(namespace string) (int, error) {
	list, err := t.Pods(namespace)

	return len(list.Items), err
}

/*Pods returns the Pods().List()  forwarded from the Clientset in the namespace
'' means the default namespace
*/
func (t *Check) Pods(namespace string) (*apiv1.PodList, error) {
	return t.Clientset.Pods(namespace).List(metav1.ListOptions{})
}
