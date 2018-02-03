package client

import (
	"errors"
	"os"
	"os/exec"
	"testing"
	"text/template"

	"github.com/stretchr/testify/assert"
)

func createRulechecker() error {
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

func deleteRulechecker() error {
	cmd := "kubectl"

	args := []string{"delete", "-f", "test-rules.yaml"}
	exec.Command(cmd, args...).Run()

	args = []string{"delete", "-f", "test-rulecheckers-rsd.yaml"}
	exec.Command(cmd, args...).Run()

	check := []string{"get", "trchk", "test-rules"}
	if err := exec.Command(cmd, check...).Run(); err != nil {
		return nil
	}

	return errors.New("remove failed, rules still exist")
}

func TestArgParseTemplate(t *testing.T) {
	assert.Nil(t, create)
	clientset, checktemplate, rules, kind, debug, err := Set([]string{"-t", "../check.tmpl"})
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
	clientset, checktemplate, rules, kind, debug, err := Set([]string{"-k", "testrulecheckers", "test-rules"})
	assert.False(t, debug)
	assert.Nil(t, checktemplate)
	assert.NotNil(t, rules)
	assert.NotNil(t, "", kind)
	assert.NotNil(t, clientset)
	assert.Nil(t, checktemplate)
	assert.Nil(t, err)
	chk := Check{Check: true, Clientset: clientset.CoreV1()}
	assert.True(t, chk.Check)
	err = listCRD(clientset, "stable.example.com", "v1", kind.(string), rules.(string))
	assert.Nil(t, err)
}

var create, delete error

func TestMain(m *testing.M) {
	create = createRulechecker()
	code := m.Run()
	delete = deleteRulechecker()
	if delete != nil {
		panic(delete)
	}
	os.Exit(code)
}
