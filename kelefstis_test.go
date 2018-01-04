package main

import (
	"errors"
	"os"
	"os/exec"
	"testing"
	"text/template"

	"github.com/endocode/kelefstis/client"
	"github.com/stretchr/testify/assert"
)

func create_rulechecker() error {
	cmd := "kubectl"

	check := []string{"get", "trchk", "test-rules"}
	if err := exec.Command(cmd, check...).Run(); err != nil {
		args := []string{"create", "-f", "test/test-rulecheckers-rsd.yaml"}
		exec.Command(cmd, args...).Run()
		args = []string{"create", "-f", "test/test-rules.yaml"}
		err = exec.Command(cmd, args...).Run()

		return err
	}

	return nil
}

func delete_rulechecker() error {
	cmd := "kubectl"

	args := []string{"delete", "-f", "test/test-rules.yaml"}
	exec.Command(cmd, args...).Run()

	args = []string{"delete", "-f", "test/test-rulecheckers-rsd.yaml"}
	exec.Command(cmd, args...).Run()

	check := []string{"get", "trchk", "test-rules"}
	if err := exec.Command(cmd, check...).Run(); err != nil {
		return nil
	} else {
		return errors.New("remove failed, rules still exist")
	}

}

func TestArgParseTemplate(t *testing.T) {
	assert.Nil(t, create_rulechecker())
	clientset, checktemplate, rules, kind, err := client.ClientSet([]string{"-t", "check.tmpl"})
	assert.Nil(t, rules)
	assert.Nil(t, kind)
	assert.NotNil(t, clientset)
	assert.NotNil(t, checktemplate)
	assert.Nil(t, err)
	chk := client.Check{Check: true, Clientset: clientset.CoreV1()}

	tmpl, err := template.New("test").Parse(checktemplate.(string))
	assert.Nil(t, err)
	assert.Nil(t, tmpl.Execute(os.Stdout, &chk))
	assert.Nil(t, delete_rulechecker())
}

func TestArgParseRules(t *testing.T) {
	assert.Nil(t, create_rulechecker())
	clientset, checktemplate, rules, kind, err := client.ClientSet([]string{"-k", "testrulecheckers", "test-rules"})
	assert.Nil(t, checktemplate)
	assert.NotNil(t, rules)
	assert.NotNil(t, "", kind)
	assert.NotNil(t, clientset)
	assert.Nil(t, checktemplate)
	assert.Nil(t, err)
	chk := client.Check{Check: true, Clientset: clientset.CoreV1()}
	assert.True(t, chk.Check)
	err = client.ListCRD(clientset, "stable.example.com", "v1", kind.(string), rules.(string))
	assert.Nil(t, err)
	assert.Nil(t, delete_rulechecker())
}
