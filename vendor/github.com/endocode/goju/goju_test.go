package goju

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func datafile(f string) string {
	return "data/" + f + ".json"
}

func TestPlayWithSimpleItemArray(t *testing.T) {
	testPodWithRules(t, "itempods", "itemrule", 2, 0, 0, 1)
}

func TestPlayWithSimplePods(t *testing.T) {
	testPodWithRules(t, "simplepods", "simplerule", 2, 0, 0, 4)
}

func TestPlayWithImagePod(t *testing.T) {
	testPodWithRules(t, "imagepod", "imagerule", 1, 0, 0, 2)
}

func TestPlayWithImageFullPod(t *testing.T) {
	testPodWithRules(t, "fullpods", "fullpodimagerule", 4, 0, 1, 30)
}

func TestPlayWithUnknownMethod(t *testing.T) {
	testPodWithRules(t, "itempods", "unknownmethodrule", 2, 1, 0, 0)
}

func TestPlay(t *testing.T) {
	assert.Nil(t, Play("data/itempods.json", "data/itemrule.json"))
}

func TestPlayFalseFirstFile(t *testing.T) {
	assert.NotNil(t, Play("data/nonono.json", "data/itemrule.json"))
}

func TestPlayFalseSecondFile(t *testing.T) {
	assert.NotNil(t, Play("data/itempods.json", "data/nonono.json"))
}

func TestPlayWithoutFile(t *testing.T) {
	var tree map[string]interface{}
	tr := &TreeCheck{Check: &Check{}}

	err := ReadFile("notexisting", &tree)
	tr.bookkeep(true, err)
	assert.NotNil(t, tr.ErrorHistory.Front())
}

func testPodWithRules(t *testing.T, treeFile, ruleFile string,
	treeLengthExpected, errorLengthExpected,
	falseExpected, trueExpected int) {
	tr := &TreeCheck{Check: &Check{}}

	var tree, ruletree map[string]interface{}
	assert.Nil(t, ReadFile(datafile(treeFile), &tree), treeFile)
	assert.Nil(t, ReadFile(datafile(ruleFile), &ruletree), ruleFile)

	assert.Len(t, tree, treeLengthExpected, "tree length")
	tr.traverse("", "/", tree, ruletree)
	assert.Equal(t, errorLengthExpected, tr.ErrorHistory.Len(), "errors")
	if errorLengthExpected == 0 {
		assert.Nil(t, tr.ErrorHistory.Front(), "error history")
	}
	assert.Equal(t, falseExpected, tr.FalseCounter, "falseCounter")
	assert.Equal(t, trueExpected, tr.TrueCounter, "trueCounter")
}

func TestPodWithWrongType(t *testing.T) {
	tr := &TreeCheck{Check: &Check{}}

	var tree, ruletree map[string]interface{}
	assert.Nil(t, ReadFile(datafile("itempods"), &tree), "wrongtype")
	assert.Nil(t, ReadFile(datafile("itemrule"), &ruletree), "wrongtype")

	tree["apiVersion"] = tr
	assert.Len(t, tree, 2, "tree length")
	tr.traverse("", "/", tree, ruletree)
	assert.Equal(t, 1, tr.ErrorHistory.Len(), "errors")
	assert.NotNil(t, tr.ErrorHistory.Front(), "error history")

	assert.Equal(t, 0, tr.FalseCounter, "falseCounter")
	assert.Equal(t, 1, tr.TrueCounter, "trueCounter")
}

func TestPodWithWrongRuleType(t *testing.T) {
	tr := &TreeCheck{Check: &Check{}}

	var tree, ruletree map[string]interface{}
	assert.Nil(t, ReadFile(datafile("itempods"), &tree), "wrongtype")
	assert.Nil(t, ReadFile(datafile("itemrule"), &ruletree), "wrongtype")
	ruletree["items"] = tr
	assert.Len(t, tree, 2, "tree length")
	tr.traverse("", "/", tree, ruletree)

	assert.Equal(t, 2, tr.ErrorHistory.Len(), "errors")
	assert.NotNil(t, tr.ErrorHistory.Front(), "error history")

	assert.Equal(t, 0, tr.FalseCounter, "falseCounter")
	assert.Equal(t, 0, tr.TrueCounter, "trueCounter")
}

func TestPlayWithWrongNumberInsteadString(t *testing.T) {
	assert.Nil(t, Play("data/fullpods.json", "data/wrongnumber.json"))
}

func TestPlayPrivilegedBool(t *testing.T) {
	assert.Nil(t, Play("data/privileged.json", "data/bool.json"))
}

func TestMain(m *testing.M) {
	code := m.Run()

	os.Exit(code)
}
