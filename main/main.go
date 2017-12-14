package main

import (
	"github.com/endocode/kelefstis/check"
	"github.com/endocode/kelefstis"
	"text/template"

	"os"
)


// simple k8s client that lists all available pods
// it gets config from $HOME/.kube/config
func main()  {

	clientset, checktemplate, err := kelefstis.ClientSet(nil)
	chk := check.Check{true, clientset.CoreV1()}
	tmpl, err := template.New("test").Parse(string(checktemplate))
	if err != nil {
		panic(err)
	}
	err = tmpl.Execute(os.Stdout, &chk)
	if err != nil {
		panic(err.Error())
	}

	kelefstis.ListNodes(clientset)
	kelefstis.ListPods(clientset)
	kelefstis.ListResource(clientset)

	kelefstis.ListCRD(clientset, "stable.example.com", "v1", "rulecheckers","rules")

}