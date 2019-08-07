package main

import (
	"github.com/golang/glog"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// Maps defines the place for all unstructured objects
type Maps struct {
	M map[string]map[string]*unstructured.Unstructured
}

func (maps *Maps) list(level glog.Level) {
	if glog.V(level) {
		for k, m := range maps.M {
			glog.Infof("unstructured %s:\n", k)
			for l := range m {
				glog.Infof("unstructured    %s", l)
			}
		}
	}
}

func (maps *Maps) store(u *unstructured.Unstructured) {
	ca := canonicalAPIVersionKind(u)
	m, ok := maps.M[ca]
	if !ok {
		m = make(map[string]*unstructured.Unstructured)
		maps.M[ca] = m
	}
	m[canonicalNamespaceName(u)] = u
}

func (maps *Maps) remove(u *unstructured.Unstructured) {
	ca := canonicalAPIVersionKind(u)
	m, ok := maps.M[ca]
	if !ok {
		return
	}
	delete(m, canonicalNamespaceName(u))
	if len(m) == 0 {
		delete(maps.M, ca)
	}
}
