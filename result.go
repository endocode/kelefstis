package main

import (
	"container/list"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"

	"github.com/golang/glog"
)

// CanonicalName is a wrapper for the string
type CanonicalName struct {
	string
}

// ResultMap is a map where the method are protected by Mutexes
type ResultMap struct {
	sync.RWMutex
	internal map[CanonicalName]*TreeCheck
}

// NewResultMap constructs a ResultMap returning the ponter
func NewResultMap() *ResultMap {
	return &ResultMap{
		internal: make(map[CanonicalName]*TreeCheck),
	}
}

// Log logs the content of the map to glog
func (rm *ResultMap) Log(l glog.Level) {
	if glog.V(l) {
		rm.RLock()
		glog.Infof("results for %d entries, %d rules", len(rm.internal), len(ruleMap))
		for k, v := range rm.internal {
			glog.Infof("%s checks true/false %d/%d", k.string, v.Result.TrueCounter, v.Result.FalseCounter)
		}
		rm.RUnlock()
	}
}

// Load returns a the value of a key safely and ok if the key has been found.
// It is protected by a read lock
func (rm *ResultMap) Load(key CanonicalName) (value *TreeCheck, ok bool) {
	rm.RLock()
	result, ok := rm.internal[key]
	rm.RUnlock()
	return result, ok
}

// Delete deletes a key safely protected by a Lock
func (rm *ResultMap) Delete(key CanonicalName) {
	rm.Lock()
	delete(rm.internal, key)
	rm.Unlock()
}

// Store stores a value
func (rm *ResultMap) Store(key CanonicalName, value *TreeCheck) {
	rm.Lock()
	rm.internal[key] = value
	rm.Unlock()
}

// Report shows the detail of a Result on glog with
// the appropriate level
func (t *Result) Report(n, k string) {
	if glog.V(2) {
		glog.Infof("checking object %s by rule %s: %d true, %d false",
			n, k, t.TrueCounter, t.FalseCounter)
		f := t.ErrorHistory.Front()
		if f != nil {
			glog.Errorf("found %d errors", t.ErrorHistory.Len())
			i := 0
			for e := f; e != nil; e = e.Next() {
				glog.Errorf("  error %d: %s", i, e.Value)
				i++
			}
		}
	}
}

// Evaluate reports true, if no errors and no false results have been found,
// and at least one true result
func (t *TreeCheck) Evaluate() bool {
	return t.Result.ErrorHistory.Len() == 0 && t.Result.FalseCounter == 0 && t.Result.TrueCounter > 0
}

// Reset sets all counters to 0 and removes the errors in the list
func (t *Result) Reset() {
	t.TrueCounter = 0
	t.FalseCounter = 0
	t.ErrorHistory.Init()
}

// Result counts the results of an Object
type Result struct {
	ErrorHistory              list.List
	TrueCounter, FalseCounter int
}

// TreeCheck is the object collection all data on a traversal
type TreeCheck struct {
	Check
	Result Result
}

// Bookkeep adds an entry to the Result,
// either a true or false value to the counter or an error
// to the ErrorHistory
func (t *Result) Bookkeep(b bool, err error) {
	if err == nil {
		if b {
			t.TrueCounter++
		} else {
			t.FalseCounter++
		}
	} else {
		errn := fmt.Errorf("error #%d: %s", t.ErrorHistory.Len(), err.Error())
		t.ErrorHistory.PushBack(errn)
	}
}

// AddError adds an error to the list of errors,
// format and args are format used to create a formatted error message
func (t *TreeCheck) AddError(format string, args ...interface{}) {
	errn := fmt.Sprintf("error #%d: ", t.Result.ErrorHistory.Len())
	glog.V(2).Infof(errn+format, args...)
	t.Result.ErrorHistory.PushBack(fmt.Errorf(errn+format, args...))
}

func (t *TreeCheck) typeCall(name string, inputs []reflect.Value) (result []reflect.Value, err error) {
	v := reflect.ValueOf(&t.Check)
	met := v.MethodByName(name)
	if met.Kind() == reflect.Invalid {
		err = errors.New("unknown method: " + name)
		return
	}
	fun := reflect.TypeOf(met.Interface())
	result = make([]reflect.Value, fun.NumIn())
	for i := 0; i < fun.NumIn(); i++ {
		if fun.In(i) == inputs[i].Type() {
			result[i] = inputs[i]
		} else {
			glog.V(4).Infof("type mismatch, need to convert from %s to %s", inputs[i].Type(), fun.In(i))
			if fun.In(i).Kind() == reflect.Int64 {
				if inputs[i].Type().Kind() == reflect.String {
					var i64 int64
					i64, err = strconv.ParseInt(inputs[i].String(), 10, 64)
					if err != nil {
						return
					}
					result[i] = reflect.ValueOf(i64)
				}
			}
		}
	}
	err = nil
	return
}

func (t *TreeCheck) nestedCheck(offset string, c, r interface{}) {
	f := reflect.ValueOf(r)
	g := reflect.ValueOf(c)
	fk := f.Kind()
	gk := g.Kind()
	glog.V(12).Infof("%s found types %s // %s", offset, gk, fk)

	switch fk {
	case reflect.Slice:
		{
			a, oka := r.([]interface{})
			for i := range a {
				ii := int(i)
				if gk == reflect.Invalid {
					t.nestedCheck(offset+"<invalid>", nil, a[ii])
				} else {
					b, okb := c.([]interface{})
					if oka && okb {

						is := fmt.Sprintf("%d", ii)
						t.nestedCheck(offset+"["+is+"]", b[ii], a[ii])
					}
				}
			}
		}
	case reflect.Map:
		{
			a, oka := r.(map[string]interface{})
			for i := range a {
				switch gk {
				case reflect.Invalid:
					{
						t.nestedCheck(offset+"/<invalid>", nil, a[i])
					}
				case reflect.Map:
					{
						b, okb := c.(map[string]interface{})
						if oka && okb {
							t.nestedCheck(offset+"/"+i, b[i], a[i])
						}
					}
				case reflect.Int64:
					{
						glog.Infof("found Int64 #######")
					}
				default:
					{
						s := reflect.Value(g).String()
						glog.V(12).Infof("checking %s: %s by rule %s:%s", offset, g, i, a[i])
						inputs := make([]reflect.Value, 2)
						inputs[1] = reflect.ValueOf(s)
						inputs[0] = reflect.ValueOf(a[i])
						n := strings.Title(i)
						if glog.V(8) {
							glog.Infof("method name=%s", n)
						}

						v := reflect.ValueOf(&t.Check)
						met := v.MethodByName(n)

						if met.Kind() == reflect.Invalid {
							glog.Errorf("invalid method %s", n)
						} else {
							m := reflect.TypeOf(met.Interface())

							glog.V(8).Infof("method = %v, input=%s:%v, %s:%v, args %v %v",
								met, inputs[0].Type(), inputs[0], inputs[1].Type(), inputs[1], m.In(0), m.In(1))
							r, err := t.typeCall(n, inputs)
							if err != nil {
								glog.Errorf("error creating args %s", err.Error())
							}
							res := met.Call(r)
							bo := res[0].Bool()
							e := "ok"
							var ee error
							if !res[1].IsNil() {
								ee = res[1].Interface().(error)
								e = ee.Error()
							}
							t.Result.Bookkeep(bo, ee)

							glog.V(8).Infof("result %s(%v,%v) = (%t, %s)", n, inputs[0], inputs[1], bo, e)

						}
					}
				}
			}
		}
	default:
		{
		}
	}

}
