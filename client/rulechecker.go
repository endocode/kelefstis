package client

/*

created with https://mholt.github.io/json-to-go/

 */
type RuleChecker struct {
	Kind string `json:"kind"`
	Spec struct {
		Rules []struct {
			Pods struct {
				Namespace struct {
					Eq string `json:"$eq"`
				} `json:"namespace"`
				Spec struct {
					Containers struct {
						Image struct {
							Matches string `json:"matches"`
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
	APIVersion string `json:"apiVersion"`
	Metadata   struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	} `json:"metadata"`
}
