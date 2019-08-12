package main

import (
	"os"
	"time"

	"github.com/ghodss/yaml"
	"github.com/golang/glog"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
)

// NewDynamicListWatchFromConfig creates a ListWatcher for dynamic, unstructured output
// not types are used
func newDynamicListWatchFromConfig(cfg *rest.Config, group, version, resource string) (*cache.ListWatch, error) {
	dynamicClient, err := dynamic.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}
	gvr := schema.GroupVersionResource{Group: group, Version: version, Resource: resource}
	list :=
		func(opts metav1.ListOptions) (runtime.Object, error) {
			return dynamicClient.Resource(gvr).List(opts)
		}
	watch :=
		func(opts metav1.ListOptions) (watch.Interface, error) {
			return dynamicClient.Resource(gvr).Watch(opts)
		}
	return &cache.ListWatch{ListFunc: list, WatchFunc: watch}, nil

}
func canonicalAPIVersionKind(u *unstructured.Unstructured) string {
	return u.GetAPIVersion() + ":" + u.GetKind()
}

func canonicalNamespaceName(u *unstructured.Unstructured) string {
	return u.GetNamespace() + "/" + u.GetName()
}

func canonicalName(u *unstructured.Unstructured) string {
	return canonicalAPIVersionKind(u) + "/" + canonicalNamespaceName(u)
}

// LogRuleChecker logs the unstructured object go glog
// with verbosity -v 8 you get also the yaml output
func LogRuleChecker(verb string, o interface{}) {
	u, ok := o.(*unstructured.Unstructured)
	if ok {
		fields := []string{"spec", "status", "metadata"}
		for _, f := range fields {
			n, _, _ := unstructured.NestedSlice(u.Object, f, "rules")
			for _, r := range n {
				i, _ := r.(map[string]interface{})
				glog.Infof("%s: %s",
					verb, canonicalName(u))

				if glog.V(12) {
					p, _, _ := unstructured.NestedFieldNoCopy(i, "pods")
					glog.Infof("%s:\n%v\n", f, p)

					b, err := yaml.Marshal(u)
					if err == nil {
						glog.Infof("%s", b)
					} else {
						glog.Errorf("Error converting unstructured to yaml %s", err.Error())
					}
				}
			}
		}
	} else {
		glog.Errorf("error handling object %s", o)
	}
}

// EventHandlers returns the functions for the event handling
func EventHandlers(log func(string, interface{})) *cache.ResourceEventHandlerFuncs {
	return &cache.ResourceEventHandlerFuncs{
		AddFunc: func(o interface{}) {
			log("add", o)
			u, _ := o.(*unstructured.Unstructured)
			n := CanonicalName{string: canonicalName(u)}
			check := &TreeCheck{Result: Result{}}

			results.Store(n, check)
			check.checkObjectByAllRules(u.Object, n.string)
		},
		DeleteFunc: func(o interface{}) {
			log("delete", o)
			n := CanonicalName{string: canonicalName(o.(*unstructured.Unstructured))}
			results.Delete(n)
		},
		UpdateFunc: func(old, new interface{}) {
			log("update old", old)
			log("update new", new)
			u := new.(*unstructured.Unstructured)
			n := CanonicalName{string: canonicalName(u)}
			check, _ := results.Load(n)
			check.Result.Reset()
			check.checkObjectByAllRules(u.Object, n.string)

			//u, _ := new.(*unstructured.Unstructured)
		},
	}
}

func (t *TreeCheck) checkObjectByAllRules(object interface{}, n string) {
	for k, r := range ruleMap {
		t.checkObjectByRule(object, r)
		t.Result.Report(n, k)
	}
}

func reportMarshal(field ...interface{}) {
	for i, f := range field {
		oy, oerr := yaml.Marshal(f)
		if oerr == nil {
			glog.Infof("%d: --------------\n%s", i, oy)
		} else {
			glog.Errorf("error %s", oerr)
		}
	}
}

func (t *TreeCheck) checkObjectByRule(object interface{}, rule *unstructured.Unstructured) {
	rules, found, err := unstructured.NestedSlice(rule.Object, "spec", "rules")
	if !found {
		glog.Errorf("could not find spec/rules in %s", canonicalName(rule))
	}
	if err != nil {
		glog.Errorf("error %s looking up %s", err.Error(), canonicalName(rule))

	}
	for i, r := range rules {
		rm := r.(map[string]interface{})
		if glog.V(12) {
			glog.Infof("part rule %d -----------", i)
			reportMarshal(object, rm)
		}
		t.nestedCheck("", object, r)
	}
}

// RuleEventHandlers returns the functions for the event handling
func RuleEventHandlers(log func(string, interface{})) *cache.ResourceEventHandlerFuncs {
	return &cache.ResourceEventHandlerFuncs{
		AddFunc: func(o interface{}) {
			log("add", o)
			u, _ := o.(*unstructured.Unstructured)
			n := canonicalName(u)
			glog.Infof("adding rule %s", n)
			ruleMap[n] = u

		},
		DeleteFunc: func(o interface{}) {
			log("delete", o)
			n := canonicalName(o.(*unstructured.Unstructured))
			delete(ruleMap, n)
		},
		UpdateFunc: func(old, new interface{}) {
			log("update old", old)
			log("update new", new)
			n := canonicalName(new.(*unstructured.Unstructured))
			//check, _ := ruleMap[n]
			u, _ := new.(*unstructured.Unstructured)
			ruleMap[n] = u
		},
	}
}

// NewController creates a controller from group, config, version
// and resource handlers
func NewController(cfg *restclient.Config,
	group, version, resource string,
	eventHandlers *cache.ResourceEventHandlerFuncs) cache.Controller {
	listwatch, err := newDynamicListWatchFromConfig(cfg, group, version, resource)
	if err != nil {
		glog.Fatalf("Error building dynamic client for %s/%s: %s", version, resource, err.Error())
	}

	_, controller := cache.NewInformer(
		listwatch,
		&unstructured.Unstructured{},
		time.Second*30,
		eventHandlers,
	)
	return controller
}

func main() {
	os.Mkdir("/tmp", 0777)
	glog.Info("Creating watch")

	eventObjectHandlers := EventHandlers(LogObject)
	eventRuleCheckerHandlers := RuleEventHandlers(LogRuleChecker)
	stop := make(chan struct{})

	ruleController := NewController(cfg, "kelefstis.endocode.com", "v1alpha1", "rulecheckers", eventRuleCheckerHandlers)

	go ruleController.Run(stop)

	podController := NewController(cfg, "", "v1", "pods", eventObjectHandlers)
	nodeController := NewController(cfg, "", "v1", "nodes", eventObjectHandlers)
	glog.Info("Creating channel")

	go podController.Run(stop)
	go nodeController.Run(stop)

	<-stop
	glog.Info("Stopping")
}
