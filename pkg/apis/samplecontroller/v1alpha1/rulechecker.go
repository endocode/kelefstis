/*
Copyright 2017 The Kubernetes Authors.

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

package v1alpha1

import (
	"encoding/json"

	"github.com/golang/glog"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// RuleChecker is a specification for a RuleChecker resource
type RuleChecker struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RuleCheckerSpec   `json:"spec"`
	Status RuleCheckerStatus `json:"status"`
}

// RuleCheckerSpec is the spec for a RuleChecker resource
type RuleCheckerSpec struct {
	Rules []RuleSpec `json:"rules"`
}

// Rulespec defines a single rule for a kubernetes type
// M will be hidden for (Un)Marshaling
type RuleSpec struct {
	M map[string]interface{} `json:",inline"`
}

// UnmarschalJSON unmarshals the []byte array into the
// M internal map effectively hiding M
func (r *RuleSpec) UnmarshalJSON(data []byte) error {
	if glog.V(8) {
		glog.Infof("%s\n", data)
	}
	return json.Unmarshal(data, &r.M)
}

// MarschalJSON marshals the RuleSpec array from the
// M internal map effectively hiding M
func (r *RuleSpec) MarshalJSON() ([]byte, error) {
	return json.Marshal(&r.M)
}

func (m *RuleSpec) DeepCopyInto(in *RuleSpec) {
	glog.Infof("#####################\ncopying %+V\n##############\n", m)
	panic(nil)
}

// RuleCheckerStatus is the status for a RuleChecker resource
type RuleCheckerStatus struct {
	Rules RuleSpec `json:"rules"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// RuleCheckerList is a list of RuleChecker resources
type RuleCheckerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []RuleChecker `json:"items"`
}
