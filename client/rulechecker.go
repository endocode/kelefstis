package client

import "time"

/*

created with https://mholt.github.io/json-to-go/

 */
type RuleChecker struct {
	APIVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
	Metadata   struct {
		ClusterName                string      `json:"clusterName"`
		CreationTimestamp          time.Time   `json:"creationTimestamp"`
		DeletionGracePeriodSeconds interface{} `json:"deletionGracePeriodSeconds"`
		DeletionTimestamp          interface{} `json:"deletionTimestamp"`
		Description                string      `json:"description"`
		Initializers               interface{} `json:"initializers"`
		Name                       string      `json:"name"`
		Namespace                  string      `json:"namespace"`
		ResourceVersion            string      `json:"resourceVersion"`
		SelfLink                   string      `json:"selfLink"`
		UID                        string      `json:"uid"`
	} `json:"metadata"`
	Spec struct {
		Rules []struct {
			Domain   string `json:"domain"`
			MinNodes int    `json:"minNodes,omitempty"`
		} `json:"rules"`
	} `json:"spec"`
}