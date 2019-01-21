package goju

import (
	"container/list"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"reflect"
	"strconv"
	"strings"

	"github.com/golang/glog"
)

// TreeCheck is the object collection all data on a traversal
type TreeCheck struct {
	Check                     *Check
	ErrorHistory              list.List
	TrueCounter, FalseCounter int
}

// AddError adds an error to the list of errors,
// format and args are format used to create a formatted error message
func (t *TreeCheck) AddError(format string, args ...interface{}) {
	errn := fmt.Sprintf("error #%d: ", t.ErrorHistory.Len())
	glog.V(2).Infof(errn+format, args...)
	t.ErrorHistory.PushBack(fmt.Errorf(errn+format, args...))
}

func (t *TreeCheck) bookkeep(b bool, err error) {
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

func cutString(i interface{}, l int) string {
	var out string
	if i == nil {
		out = "nada"
	} else {
		out = fmt.Sprintf("%s", i)
	}
	if len(out) > l {
		return out[0:l] + " ..."
	}
	return out
}

// ToStringValue turns the input interface into a Value based
// on String
func ToStringValue(i interface{}) reflect.Value {
	f := reflect.ValueOf(i)
	switch f.Kind() {

	case reflect.Float64:
		return reflect.ValueOf(strconv.FormatFloat(f.Float(), 'g', -1, 64))
	case reflect.Bool:
		if f.Bool() {
			return reflect.ValueOf("true")
		}
		return reflect.ValueOf("false")
	case reflect.String:
		return f
	}
	return f
}

func (t *TreeCheck) methodCall(capMethod, offset, path string, v, tv interface{}) interface{} {
	method := reflect.ValueOf(t.Check).MethodByName(capMethod)
	if method.IsValid() {
		glog.V(5).Infof("%s\t rules %s %s %s ", offset, capMethod, v, cutString(tv, 40))

		conv := ToStringValue(v)
		tvconv := ToStringValue(tv)

		result := method.Call([]reflect.Value{conv, tvconv})

		ok := result[0].Bool()
		err := result[1].Interface()
		level := 2
		if ok {
			t.TrueCounter++
		} else {
			t.FalseCounter++
			level = 1
		}
		threshold := glog.Level(level)
		if glog.V(threshold) {
			glog.V(threshold).Infof("#%d: %s.%s(%q,%q): %t",
				t.TrueCounter+t.FalseCounter, path, capMethod, conv, tvconv, ok)
		}
		return err
	}

	msg := fmt.Sprintf("no method %q", capMethod)
	glog.V(5).Infof(msg)

	return errors.New(msg)
}

func (t *TreeCheck) applyRule(offset, path string, treeValue reflect.Value,
	rulesValue reflect.Value, rules interface{}) {
	glog.V(5).Info(offset, "\t rules value Kind ", rulesValue.Kind())
	switch rulesValue.Kind() {
	case reflect.Map, reflect.String:
		m, ok := rules.(map[string]interface{})
		if ok {
			tv := treeValue.Interface()

			for k, v := range m {
				capMethod := strings.Title(k)
				err := t.methodCall(capMethod, offset, path, v, tv)
				if err != nil {
					switch treeValue.Kind() {
					case reflect.String, reflect.Float64, reflect.Bool:
						{
							e, _ := err.(error)
							t.AddError(e.Error())
						}
					}
				}
			}
		}
	default:
		{
			t.AddError("found unknown ruleValue %q with value %q", rulesValue.Kind(), rulesValue)
		}
	}
	//	fmt.Printf("# errors %d %d\n", t.falseCounter, t.trueCounter)
}

// Traverse check if tree complies according to rules
// Both are dictionaries with strings as keys
// and dictionaries or strings as value
func (t *TreeCheck) Traverse(tree interface{}, rules interface{}) {
	t.traverse("", "", tree, rules)
}

func (t *TreeCheck) traverse(offset, path string, tree interface{}, rules interface{}) {
	if tree == nil || rules == nil {
		glog.V(5).Infof(offset+"< traverse t is nil=%t r is nil=%t>\n", tree == nil, rules == nil)
		return
	}
	treeValue := reflect.ValueOf(tree)
	rulesValue := reflect.ValueOf(rules)
	glog.V(5).Info(offset, "< traverse", treeValue.Type())

	switch treeValue.Kind() {

	case reflect.Slice, reflect.Array:
		t.applyRule(offset, path, treeValue, rulesValue, rules)
		ti, ok := tree.([]interface{})
		if ok {
			for i, vi := range ti {
				index := fmt.Sprintf("%d:", i)
				index = ""
				t.traverse(offset+index+"\t", fmt.Sprintf(".%s[%d]", path, +i), vi, rules)
			}
		}

	case reflect.Map:
		for k, v := range tree.(map[string]interface{}) {
			r, ok := rulesValue.Interface().(map[string]interface{})
			if ok {
				// fmt.Printf("### ok key %q %v =: %q \n", k, cutString(v, 30), cutString(r[k], 30))
				t.traverse(offset+"\t ", fmt.Sprintf("%s.%s", path, k), v, r[k])
			} else {
				// fmt.Printf("#### not ok")
				t.applyRule(offset, path, treeValue, rulesValue, r)
			}
		}

	case reflect.String, reflect.Float64, reflect.Bool:
		t.applyRule(offset, path, treeValue, rulesValue, rules)
	default:
		glog.V(5).Info(" == unknown ", treeValue)
		t.AddError("found unknown type %v with value %q", treeValue, treeValue)
	}
	glog.V(5).Info(offset, ">")
}

//ReadFile reads file f and unmarshal it into t, reporting the error
func ReadFile(f string, t interface{}) error {
	b, err := ioutil.ReadFile(f)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, &t)
}

// Play calls traverse check a json files by the rules in the second json file
func Play(json, rule string) error {
	//usage := fmt.Sprintf("usage: %s [options] <data> <rules>\n\noptions are:\n\n", os.Args[0])

	var tree, ruletree map[string]interface{}
	err := ReadFile(json, &tree)
	if err != nil {
		return err
	}
	err = ReadFile(rule, &ruletree)
	if err != nil {
		return err
	}

	tr := &TreeCheck{Check: &Check{}}

	tr.Traverse(tree, ruletree)

	glog.V(1).Infof("Errors       : %d\n", tr.ErrorHistory.Len())
	glog.V(1).Infof("Checks   true: %d\n", tr.TrueCounter)
	glog.V(1).Infof("Checks  false: %d\n", tr.FalseCounter)

	return nil
}
