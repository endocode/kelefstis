package test


import (
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/endocode/kelefstis/client"
	"os/exec"
	"errors"
	"text/template"
	"os"
)

func create_rulechecker()(error){
	cmd := "kubectl"

	check :=  []string{"get","trchk","test-rules"}
	if err := exec.Command(cmd, check...).Run(); err != nil {
		args := []string{"create","-f","test-rulecheckers-rsd.yaml"}
		exec.Command(cmd, args...).Run()
		args = []string{"create","-f","test-rules.yaml"}
		err= exec.Command(cmd, args...).Run();

		return err
	}

	return nil
}

func delete_rulechecker()(error){
	cmd := "kubectl"

		args := []string{"delete","-f","test-rules.yaml"}
		exec.Command(cmd, args...).Run()

		args = []string{"delete","-f","test-rulecheckers-rsd.yaml"}
		exec.Command(cmd, args...).Run();

	check :=  []string{"get","trchk","test-rules"}
	if err := exec.Command(cmd, check...).Run(); 	err!=nil {
		return nil
	} else {
		return errors.New("remove failed, rules still exist")
	}

}

func TestArgParse(t *testing.T) {
	assert.Nil(t,create_rulechecker())
	clientset, checktemplate, err := client.ClientSet([]string{"../check.templ"})
	assert.NotNil(t, clientset)
	assert.NotNil(t,checktemplate)
	assert.Nil(t,err)
	chk := client.Check{true, clientset.CoreV1()}

	tmpl, err := template.New("test").Parse(string(checktemplate))
	assert.Nil(t,err)
	assert.Nil(t,tmpl.Execute(os.Stdout,&chk))
	assert.Nil(t,delete_rulechecker())
}
