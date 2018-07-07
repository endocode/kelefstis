package main

import (
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"testing"

	"github.com/endocode/goju"
	"github.com/golang/glog"
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
	tree, err := CRD(clientset, "apis", "stable.example.com", "v1", "testrulecheckers", "", "rules")
	assert.NotNil(t, tree)
	assert.Nil(t, err)

	pods, ok := toStringMap(tree["pods"])
	assert.NotNil(t, pods)
	assert.True(t, ok)

	raw, err := json.Marshal(pods)
	LogJSON(2, raw)
}

// TestRules just looks up one object
func TestPods(t *testing.T) {
	tree, err := Items(clientset, "api", "", "v1", "pods", "")
	assert.NotNil(t, tree)
	assert.Nil(t, err)

	pods, ok := toStringMap(tree[0])
	assert.NotNil(t, pods)
	assert.True(t, ok)

	raw, err := json.Marshal(pods)
	LogJSON(2, raw)
}

func TestTraverse(t *testing.T) {
	rulemap, err := CRD(clientset, "apis", "stable.example.com", "v1", "testrulecheckers", "", "rules")
	assert.Nil(t, err)
	assert.NotNil(t, rulemap)

	items, err := Items(clientset, "api", "", "v1", "pods", "")
	assert.Nil(t, err)

	assert.NotNil(t, items)

	pods, ok := toStringMap(items[0])
	assert.NotNil(t, pods)
	assert.True(t, ok)

	var treecheck = goju.TreeCheck{Check: &goju.Check{}}

	s, err := map2string(items[0])
	assert.Nil(t, err)
	assert.NotEqual(t, s, "")

	podrule, ok := toStringMap(rulemap["pods"])
	assert.NotNil(t, podrule)
	assert.True(t, ok)

	tr, err := map2string(podrule)
	assert.Nil(t, err)
	assert.NotEqual(t, tr, "")

	glog.V(0).Infof("\nTree:%s\n#################################\nRules%s", s, tr)

	treecheck.Traverse(items[0], podrule)

	for i := treecheck.Check.ErrorHistory.Front(); i != nil; i = i.Next() {
		glog.V(0).Infof("error %s", i)
	}
	assert.True(t, treecheck.Check.TrueCounter > 0)
	glog.V(0).Infof("tests errors/true/false: %d/%d/%d",
		treecheck.Check.ErrorHistory.Len(),
		treecheck.Check.TrueCounter, treecheck.Check.FalseCounter)

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
