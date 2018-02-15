package client

import (
	"fmt"
	"strings"
	"time"
	//	"fmt"
)

/*The RuleChecker holds all the data to check rules

created with https://mholt.github.io/json-to-go/

*/
type RuleChecker struct {
	APIVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
	Metadata   struct {
		ClusterName       string    `json:"clusterName"`
		CreationTimestamp time.Time `json:"creationTimestamp"`
		Description       string    `json:"description"`
		Name              string    `json:"name"`
		Namespace         string    `json:"namespace"`
		ResourceVersion   string    `json:"resourceVersion"`
		SelfLink          string    `json:"selfLink"`
		UID               string    `json:"uid"`
	} `json:"metadata"`
	Spec struct {
		Rules []struct {
			Pods struct {
				Range string `json:"$range"`
				// Namespace struct {
				// 	Eq string `json:"$eq"`
				// } `json:"namespace"`
				Spec struct {
					Containers struct {
						Range string `json:"$range"`
						Image struct {
							MatchString string `json:"$matches"`
						} `json:"image"`
					} `json:"containers"`
				} `json:"spec"`
			} `json:"pods,omitempty"`
			Cluster struct {
				Max int `json:"max"`
				Min int `json:"min"`
			} `json:"cluster,omitempty"`
			// Nodes struct {
			// 	Range  string `json:"$range"`
			// 	Memory struct {
			// 		Min string `json:"min"`
			// 	} `json:"memory"`
			// } `json:"nodes,omitempty"`
		} `json:"rules"`
	} `json:"spec"`
}

//Time for the RFC3339 marshalling
type Time struct {
	time.Time
}

//String returns Time as RFC3339 string
func (t *Time) String() string {
	return t.Time.Format(time.RFC3339)
}

// UnmarshalJSON unmarshals the Time struct
//from a RFC3339 byte buffer and handles null values
func (t *Time) UnmarshalJSON(buf []byte) error {
	b := string(buf)
	if b == "null" {
		t.Time = time.Time{}
	} else {
		s := strings.Trim(b, `"`)
		tt, err := time.Parse(time.RFC3339, s)
		if err != nil {
			fmt.Printf("Error %s\n", err)
			return err
		}
		t.Time = tt
	}
	return nil
}

//MarshalJSON converts Time to a RFC3339 byte arry
func (t *Time) MarshalJSON() ([]byte, error) {
	return []byte(`"` + t.Time.Format(time.RFC3339) + `"`), nil
}
