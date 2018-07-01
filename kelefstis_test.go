package main

import (
	"errors"
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
)

func init() {
	initFlags()
}

func createRulechecker() error {
	cmd := "kubectl"

	check := []string{"get", "trchk", "test-rules"}
	if err := exec.Command(cmd, check...).Run(); err != nil {
		args := []string{"create", "-f", "client/test-rulecheckers-rsd.yaml"}
		exec.Command(cmd, args...).Run()
		args = []string{"create", "-f", "client/test-rules.yaml"}
		err = exec.Command(cmd, args...).Run()

		return err
	}

	return nil
}

func deleteRulechecker() error {
	cmd := "kubectl"

	args := []string{"delete", "-f", "client/test-rules.yaml"}
	exec.Command(cmd, args...).Run()

	args = []string{"delete", "-f", "client/test-rulecheckers-rsd.yaml"}
	exec.Command(cmd, args...).Run()

	check := []string{"get", "trchk", "test-rules"}
	if err := exec.Command(cmd, check...).Run(); err != nil {
		return nil
	}

	return errors.New("remove failed, rules still exist")
}

// TestRules just looks up one object
func TestRules(t *testing.T) {
	rules0, err := rules(clientset, "apis", "stable.example.com", "v1", "testrulecheckers", "test-rules")
	assert.NotNil(t, rules0)
	assert.Nil(t, err)
}

var create, delete error

//TestMain is responsible to set up a test environment
func TestMain(m *testing.M) {
	create = createRulechecker()
	code := m.Run()
	delete = deleteRulechecker()
	if delete != nil {
		panic(delete)
	}
	os.Exit(code)
}
