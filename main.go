package main

import (
	"github.com/endocode/kelefstis/client"
	"text/template"
	"os"
)


// simple k8s client that lists all available pods
// it gets config from $HOME/.kube/config
func main()  {

	clientset, checktemplate, err := client.ClientSet(nil)
	chk := client.Check{true, clientset.CoreV1()}
	tmpl, err := template.New("test").Parse(string(checktemplate))
	if err != nil {
		panic(err)
	}
	err = tmpl.Execute(os.Stdout, &chk)
	if err != nil {
		panic(err.Error())
	}

	/*
	client.ListNodes(clientset)
	client.ListPods(clientset)
	client.ListResource(clientset)
	*/

//	client.ListCRD(clientset, "stable.example.com", "v1", "rulecheckers","rules")

}