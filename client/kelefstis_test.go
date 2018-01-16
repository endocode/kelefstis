package client

import (
	"errors"
	"os"
	"os/exec"
	"testing"
	"text/template"

	"github.com/stretchr/testify/assert"
)

func create_rulechecker() error {
	cmd := "kubectl"

	check := []string{"get", "trchk", "test-rules"}
	if err := exec.Command(cmd, check...).Run(); err != nil {
		args := []string{"create", "-f", "test-rulecheckers-rsd.yaml"}
		exec.Command(cmd, args...).Run()
		args = []string{"create", "-f", "test-rules.yaml"}
		err = exec.Command(cmd, args...).Run()

		return err
	}

	return nil
}

func delete_rulechecker() error {
	cmd := "kubectl"

	args := []string{"delete", "-f", "test-rules.yaml"}
	exec.Command(cmd, args...).Run()

	args = []string{"delete", "-f", "test-rulecheckers-rsd.yaml"}
	exec.Command(cmd, args...).Run()

	check := []string{"get", "trchk", "test-rules"}
	if err := exec.Command(cmd, check...).Run(); err != nil {
		return nil
	} else {
		return errors.New("remove failed, rules still exist")
	}

}

func TestArgParseTemplate(t *testing.T) {
	assert.Nil(t, create)
	clientset, checktemplate, rules, kind, debug, err := ClientSet([]string{"-t", "../check.tmpl"})
	assert.False(t, debug)
	assert.Nil(t, rules)
	assert.Nil(t, kind)
	assert.NotNil(t, clientset)
	assert.NotNil(t, checktemplate)
	assert.Nil(t, err)
	chk := Check{Check: true, Clientset: clientset.CoreV1()}

	tmpl, err := template.New("test").Parse(checktemplate.(string))
	assert.Nil(t, err)
	assert.Nil(t, tmpl.Execute(os.Stdout, &chk))
}

func TestArgParseRules(t *testing.T) {
	assert.Nil(t, create)
	clientset, checktemplate, rules, kind, debug, err := ClientSet([]string{"-k", "testrulecheckers", "test-rules"})
	assert.False(t, debug)
	assert.Nil(t, checktemplate)
	assert.NotNil(t, rules)
	assert.NotNil(t, "", kind)
	assert.NotNil(t, clientset)
	assert.Nil(t, checktemplate)
	assert.Nil(t, err)
	chk := Check{Check: true, Clientset: clientset.CoreV1()}
	assert.True(t, chk.Check)
	err = ListCRD(clientset, "stable.example.com", "v1", kind.(string), rules.(string))
	assert.Nil(t, err)
}

var create, delete error

func TestMain(m *testing.M) {
	create = create_rulechecker()
	code := m.Run()
	delete = delete_rulechecker()
	if delete != nil {
		panic(delete)
	}
	os.Exit(code)
}
