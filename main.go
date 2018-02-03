package main

import (
	"os"
	"text/template"

	"github.com/endocode/kelefstis/client"
)

// simple k8s client that lists all available pods
// it gets config from $HOME/.kube/config
func main() {

	clientset, checktemplate, rules, kind, _, err := client.Set(nil)
	if err != nil {
		panic(err)
	}
	if checktemplate != nil {
		chk := client.Check{Check: true, Clientset: clientset.CoreV1()}
		tmpl, err := template.New("test").Parse(checktemplate.(string))
		if err != nil {
			panic(err)
		}
		err = tmpl.Execute(os.Stdout, &chk)
		if err != nil {
			panic(err.Error())
		}
	} else {
		client.ListCRD(clientset, "stable.example.com", "v1", kind.(string), rules.(string))
	}

}
