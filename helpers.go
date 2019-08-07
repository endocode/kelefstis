package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"reflect"
	"strconv"
)

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

//ReadFile reads file f and unmarshal it into t, reporting the error
func ReadFile(f string, t interface{}) error {
	b, err := ioutil.ReadFile(f)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, &t)
}
