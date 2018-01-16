package client

import (
	"log"
	"os"
	"regexp"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	clientv1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

type Check struct {
	Check     bool
	Clientset clientv1.CoreV1Interface
}

func init() {
	log.SetOutput(os.Stderr)
}

func (this *Check) NotZero(i int) *Check {
	this.Check = this.Check && i != 0
	return this
}

func (this *Check) GE(n int, m int) *Check {
	this.Check = this.Check && n > m
	return this
}

func (this *Check) MatchString(r string, s string) *Check {
	match, _ := regexp.MatchString(r, s)
	log.Printf("%s %s %b", r, s, match)

	this.Check = this.Check && match
	return this
}

func (this *Check) Nodes() (*apiv1.NodeList, error) {
	return this.Clientset.Nodes().List(metav1.ListOptions{})
}

func (this *Check) NumberOfNodes() (int, error) {
	list, err := this.Nodes()

	return len(list.Items), err
}

func (this *Check) NumberOfPods(namespace string) (int, error) {
	list, err := this.Pods(namespace)

	return len(list.Items), err
}

func (this *Check) Pods(namespace string) (*apiv1.PodList, error) {
	return this.Clientset.Pods(namespace).List(metav1.ListOptions{})
}
