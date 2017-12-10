package check

import "regexp"

import (
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientv1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"github.com/sirupsen/logrus"
	"os"
)

func init() {
	log.Out = os.Stdout
	log.Formatter=&logrus.JSONFormatter{}
	log.Level=logrus.FatalLevel
}

var log = logrus.New()



type Check struct {
	Check     bool
	Clientset clientv1.CoreV1Interface
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
	if log.Level >= logrus.InfoLevel {
		log.WithFields(logrus.Fields{
			"regexp": r,
			"string": s,
		}).Info("matched:")
	}
	if ! match {
		log.WithFields(logrus.Fields{
			"regexp": r,
			"string": s,
		}).Error("failed")
	}
	this.Check = this.Check && match
	return this
}

func (this *Check) Nodes() (*apiv1.NodeList, error) {
	return this.Clientset.Nodes().List(metav1.ListOptions{})
}

func (this *Check) Pods(namespace string) (*apiv1.PodList, error) {
	return this.Clientset.Pods(namespace).List(metav1.ListOptions{})
}
