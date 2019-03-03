/*
Copyright 2018 Endocode AG.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/endocode/goju"
	"github.com/golang/glog"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	appsinformers "k8s.io/client-go/informers/apps/v1"
	coreinformers "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"

	samplev1alpha1 "github.com/endocode/kelefstis/pkg/apis/kelefstis/v1alpha1"
	clientset "github.com/endocode/kelefstis/pkg/client/clientset/versioned"
	samplescheme "github.com/endocode/kelefstis/pkg/client/clientset/versioned/scheme"
	informers "github.com/endocode/kelefstis/pkg/client/informers/externalversions/kelefstis/v1alpha1"
	listers "github.com/endocode/kelefstis/pkg/client/listers/kelefstis/v1alpha1"
	"github.com/endocode/kelefstis/pkg/signals"
	"github.com/ghodss/yaml"
)

const controllerAgentName = "sample-controller"

// Controller is the controller implementation for all resources
type Controller struct {
	// kubeclientset is a standard kubernetes clientset
	kubeclientset kubernetes.Interface
	// sampleclientset is a clientset for our own API group
	sampleclientset clientset.Interface

	podsLister         corelisters.PodLister
	podsSynced         cache.InformerSynced
	ruleCheckersLister listers.RuleCheckerLister
	ruleCheckersSynced cache.InformerSynced
	informers          map[string]interface{}

	// workqueue is a rate limited work queue. This is used to queue work to be
	// processed instead of performing it as soon as a change happens. This
	// means we can ensure we only process a fixed amount of resources at a
	// time, and makes it easy to ensure we are never processing the same item
	// simultaneously in two different workers.
	workqueue workqueue.RateLimitingInterface
	// recorder is an event recorder for recording Event resources to the
	// Kubernetes API.
	recorder record.EventRecorder

	// signals so we handle the first shutdown signal gracefully
	stopCh <-chan struct{}
}

// LogJSON adds structured JSON logging to the Verbose type
func LogJSON(level glog.Level, raw []byte) error {
	if glog.V(level) {

		prettyJSON := func(raw []byte) (string, error) {
			var buffer bytes.Buffer
			err := json.Indent(&buffer, raw, "", "  ")
			return buffer.String(), err
		}
		json, err := prettyJSON(raw)
		if err != nil {
			glog.V(level).Infof("cannot convert to JSON %s", err)
			return err
		}

		lines := strings.Split(json, "\n")
		size := 1 + int(math.Log10(float64(len(lines))))
		format := "%" + strconv.Itoa(size) + "d: %s"
		for i, line := range lines {
			glog.V(level).Infof(format, i, line)
		}
	}
	return nil
}

func reUnMarshal(i interface{}) (interface{}, error) {
	b, err := yaml.Marshal(i)
	if err != nil {
		return nil, err
	}
	glog.V(12).Infof("marshaled to ------------\n%s\n------------\n")
	var r interface{}
	err = yaml.Unmarshal(b, &r)
	return r, err
}

// NewController returns a new sample controller
func NewController(
	kubeclientset kubernetes.Interface,
	sampleclientset clientset.Interface,
	deploymentInformer appsinformers.DeploymentInformer,
	podInformer coreinformers.PodInformer,
	ruleCheckersInformer informers.RuleCheckerInformer) *Controller {

	// Create event broadcaster
	// Add sample-controller types to the default Kubernetes Scheme so Events can be
	// logged for sample-controller types.
	utilruntime.Must(samplescheme.AddToScheme(scheme.Scheme))
	glog.V(4).Info("Creating event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(glog.Infof)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeclientset.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: controllerAgentName})

	controller := &Controller{
		kubeclientset:      kubeclientset,
		sampleclientset:    sampleclientset,
		podsLister:         podInformer.Lister(),
		podsSynced:         podInformer.Informer().HasSynced,
		ruleCheckersLister: ruleCheckersInformer.Lister(),
		ruleCheckersSynced: ruleCheckersInformer.Informer().HasSynced,
		workqueue:          workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "Rules"),
		recorder:           recorder,
		stopCh:             signals.SetupSignalHandler(),
		informers:          make(map[string]interface{}),
	}

	glog.Info("Setting up event handlers")
	// Set up an event handler for when rule and pod resources change
	deploymentInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.handleObject,
		UpdateFunc: func(old, new interface{}) {
			newDepl := new.(*appsv1.Deployment)
			oldDepl := old.(*appsv1.Deployment)
			if newDepl.ResourceVersion == oldDepl.ResourceVersion {
				// Periodic resync will send update events for all known Deployments.
				// Two different versions of the same Deployment will always have different RVs.
				return
			}
			controller.handleObject(new)
		},
		DeleteFunc: controller.handleObject,
	})

	checkRules := func(rules []*samplev1alpha1.RuleChecker, pods []*corev1.Pod, c string) {
		treeCheck := ReportTreeCheck{goju.TreeCheck{Check: &goju.Check{}}}

		_, ok := controller.informers["pods"]
		if !ok {
			// Wait for the caches to be synced before starting workers
			glog.Infof("Waiting for pod caches to sync, %d messages pending", len(controller.stopCh))
			if ok := cache.WaitForCacheSync(controller.stopCh, controller.podsSynced); ok {
				controller.informers["pods"] = podInformer
			} else {
				glog.Info("failed to wait for caches to sync")
				return
			}
		}
		for _, rule := range rules {

			for ri, r := range rule.Spec.Rules {
				intfb, _ := reUnMarshal(r.M["pods"])
				b, _ := yaml.Marshal(intfb)
				glog.V(4).Infof("RuleChecker %s: \n%s", c, b)

				for pi, p := range pods {
					intfp, _ := reUnMarshal(p)
					glog.V(8).Infof("rule %d, pod #%d\n", ri, pi)

					fullpath := fmt.Sprintf("Pod:v1:%s:%s: ", p.Namespace, p.Name)
					treeCheck.Traverse(fullpath, intfp, intfb)
				}
			}
		}
		treeCheck.Report(4)
	}

	checkAllPods := func(obj interface{}, note string) {
		l, err := podInformer.Lister().List(labels.Everything())

		if err == nil {
			r, ok := obj.(*samplev1alpha1.RuleChecker)
			if ok {
				checkRules([]*samplev1alpha1.RuleChecker{r}, l, note)
			} else {
				glog.V(1).Infof("this is not a pod %s ", obj)
			}
		} else {
			glog.V(1).Infof("could not list pods error %s ", err)
		}
	}

	ruleCheckersInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{

		AddFunc: func(obj interface{}) {
			checkAllPods(obj, "created")
		},

		UpdateFunc: func(obj, new interface{}) {
			//checkAllPods(new, "changed")

			n, ok1 := new.(*samplev1alpha1.RuleChecker)
			o, ok2 := obj.(*samplev1alpha1.RuleChecker)
			if ok1 && ok2 && !reflect.DeepEqual(o.Spec, n.Spec) {
				checkAllPods(new, "changed")
			}
		},

		DeleteFunc: func(obj interface{}) {
			rc := obj.(*samplev1alpha1.RuleChecker)
			b, _ := yaml.Marshal(rc.Spec.Rules)
			glog.Infof("RuleChecker deleted:  \n%s", b)
		},
	})

	report := func(c string, obj interface{}) {
		p := obj.(*corev1.Pod)
		glog.Infof("Pod %s %s: \n", p.Name, c)
		if glog.V(4) {
			b, _ := yaml.Marshal(p)
			glog.Infof("%s", b)
		}
	}

	podInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			p, ok := obj.(*corev1.Pod)
			r, err := ruleCheckersInformer.Lister().List(labels.Everything())
			if err == nil && ok {
				checkRules(r, []*corev1.Pod{p}, "created")
			} else {
				report("created", obj)
			}
		},
		UpdateFunc: func(obj, new interface{}) {
			n, ok1 := new.(*corev1.Pod)
			o, ok2 := obj.(*corev1.Pod)
			r, err := ruleCheckersInformer.Lister().List(labels.Everything())
			if err == nil && ok1 && ok2 && !reflect.DeepEqual(o.Spec, n.Spec) {
				checkRules(r, []*corev1.Pod{n}, "updated")
			} else {
				if glog.V(4) {
					report("updated", new)
				}
			}
		},
		DeleteFunc: func(obj interface{}) {
			report("deleted", obj)
		},
	})

	return controller
}

// Run will set up the event handlers for types we are interested in, as well
// as syncing informer caches and starting workers. It will block until stopCh
// is closed, at which point it will shutdown the workqueue and wait for
// workers to finish processing their current work items.
func (c *Controller) Run(threadiness int, stopCh <-chan struct{}) error {
	defer runtime.HandleCrash()
	defer c.workqueue.ShutDown()
	c.stopCh = stopCh
	// Start the informer factories to begin populating the informer caches
	glog.Info("Starting controller")

	glog.Info("Waiting for rulecheckers caches to sync")
	if ok := cache.WaitForCacheSync(stopCh, c.ruleCheckersSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	glog.Info("Starting workers")
	// Launch two workers to process resources
	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	glog.Info("Started workers")
	<-stopCh
	glog.Info("Shutting down workers")

	return nil
}

// runWorker is a long-running function that will continually call the
// processNextWorkItem function in order to read and process a message on the
// workqueue.
func (c *Controller) runWorker() {
	for c.processNextWorkItem() {
	}
}

// processNextWorkItem will read a single work item off the workqueue and
// attempt to process it, by calling the syncHandler.
func (c *Controller) processNextWorkItem() bool {
	obj, shutdown := c.workqueue.Get()

	if shutdown {
		return false
	}

	// We wrap this block in a func so we can defer c.workqueue.Done.
	err := func(obj interface{}) error {
		// We call Done here so the workqueue knows we have finished
		// processing this item. We also must remember to call Forget if we
		// do not want this work item being re-queued. For example, we do
		// not call Forget if a transient error occurs, instead the item is
		// put back on the workqueue and attempted again after a back-off
		// period.
		defer c.workqueue.Done(obj)
		var key string
		var ok bool
		// We expect strings to come off the workqueue. These are of the
		// form namespace/name. We do this as the delayed nature of the
		// workqueue means the items in the informer cache may actually be
		// more up to date that when the item was initially put onto the
		// workqueue.
		if key, ok = obj.(string); !ok {
			// As the item in the workqueue is actually invalid, we call
			// Forget here else we'd go into a loop of attempting to
			// process a work item that is invalid.
			c.workqueue.Forget(obj)
			runtime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
			return nil
		}
		// Run the syncHandler, passing it the namespace/name string of the
		// Foo resource to be synced.
		if err := c.syncHandler(key); err != nil {
			return fmt.Errorf("error syncing '%s': %s", key, err.Error())
		}
		// Finally, if no error occurs we Forget this item so it does not
		// get queued again until another change happens.
		c.workqueue.Forget(obj)
		glog.Infof("Successfully synced '%s'", key)
		return nil
	}(obj)

	if err != nil {
		runtime.HandleError(err)
		return true
	}

	return true
}

// syncHandler compares the actual state with the desired, and attempts to
// converge the two. It then updates the Status block of the Foo resource
// with the current status of the resource.
func (c *Controller) syncHandler(key string) error {
	// Convert the namespace/name string into a distinct namespace and name
	_, _, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		runtime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}
	return nil
}

// handleObject will take any resource implementing metav1.Object and attempt
// to find the Foo resource that 'owns' it. It does this by looking at the
// objects metadata.ownerReferences field for an appropriate OwnerReference.
// It then enqueues that Foo resource to be processed. If the object does not
// have an appropriate OwnerReference, it will simply be skipped.
func (c *Controller) handleObject(obj interface{}) {
	var object metav1.Object
	var ok bool
	if object, ok = obj.(metav1.Object); !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			runtime.HandleError(fmt.Errorf("error decoding object, invalid type"))
			return
		}
		object, ok = tombstone.Obj.(metav1.Object)
		if !ok {
			runtime.HandleError(fmt.Errorf("error decoding object tombstone, invalid type"))
			return
		}
		glog.V(4).Infof("Recovered deleted object '%s' from tombstone", object.GetName())
	}
	glog.V(4).Infof("Processing object: %s", object.GetName())
	if ownerRef := metav1.GetControllerOf(object); ownerRef != nil {
		// If this object is not owned by a Foo, we should not do anything more
		// with it.
		if ownerRef.Kind != "Foo" {
			return
		}
		return
	}
}
