package main

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/ghodss/yaml"
	"github.com/golang/glog"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func (t *TreeCheck) methodCall(capMethod, offset, path string, v, tv interface{}) interface{} {
	method := reflect.ValueOf(t.Check).MethodByName(capMethod)
	if method.IsValid() {
		glog.V(5).Infof("%s\t rules %s %s %s ", offset, capMethod, v, cutString(tv, 40))

		conv := ToStringValue(v)
		tvconv := ToStringValue(tv)

		result := method.Call([]reflect.Value{conv, tvconv})

		ok := result[0].Bool()
		err := result[1].Interface()
		t.Result.Bookkeep(ok, err.(error))
		if glog.V(8) {
			glog.Infof("#%d: %s.%s(%q,%q): %t",
				t.Result.TrueCounter+t.Result.FalseCounter, path, capMethod, conv, tvconv, ok)
		}
		return err
	}

	msg := fmt.Sprintf("no method %q", capMethod)
	glog.V(12).Infof(msg)

	return errors.New(msg)
}

// LogObject logs the unstructured object go glog
// with verbosity -v 8 you get also the yaml output
func LogObject(verb string, o interface{}) {
	u, ok := o.(*unstructured.Unstructured)
	if ok {
		//n, _, _ := unstructured.NestedFieldNoCopy(u.Object, "spec")
		glog.Infof("%s: %s\n",
			verb, canonicalName(u))
		if glog.V(12) {
			b, err := yaml.Marshal(u)
			if err == nil {
				glog.Infof("%s", b)
			} else {
				glog.Errorf("Error converting unstructured to yaml %s", err.Error())
			}
		}
	} else {
		glog.Errorf("error handling object %s", o)
	}
}
