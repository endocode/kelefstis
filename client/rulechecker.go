package client

import (
	"strings"
	"time"
	//	"fmt"
)

/*

created with https://mholt.github.io/json-to-go/

*/
type RuleChecker struct {
	APIVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
	Metadata   struct {
		ClusterName                string      `json:"clusterName"`
		CreationTimestamp          Time        `json:"creationTimestamp"`
		DeletionGracePeriodSeconds int         `json:"deletionGracePeriodSeconds"`
		DeletionTimestamp          Time        `json:"deletionTimestamp"`
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
			Pods struct {
				Namespace struct {
					Eq string `json:"$eq"`
				} `json:"namespace"`
				Spec struct {
					Containers struct {
						Image struct {
							Matches string `json:"$matches"`
						} `json:"image"`
					} `json:"containers"`
				} `json:"spec"`
			} `json:"pods,omitempty"`
			Cluster struct {
				Max int `json:"max"`
				Min int `json:"min"`
			} `json:"cluster,omitempty"`
			Nodes struct {
				Memory struct {
					Min string `json:"min"`
				} `json:"memory"`
			} `json:"nodes,omitempty"`
		} `json:"rules"`
	} `json:"spec"`
}

type Time struct {
	time.Time
}

func (t *Time) String() string {
	return t.Time.Format(time.RFC3339)
}

func (t *Time) UnmarshalJSON(buf []byte) error {
	s := strings.Trim(string(buf), `"`)
	//fmt.Printf("Unmarshal %s\n",s)
	tt, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return err
	}
	t.Time = tt
	return nil
}

func (t Time) MarshalJSON() ([]byte, error) {
	return []byte(`"` + t.Time.Format(time.RFC3339) + `"`), nil
}
