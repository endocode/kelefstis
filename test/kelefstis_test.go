package test


import (
"testing"
"github.com/stretchr/testify/assert"

	"github.com/endocode/kelefstis"
)

func TestSomething(t *testing.T) {
	clientset, checktemplate, err := kelefstis.ClientSet([]string{"../check.templ"})
	assert.NotNil(t, clientset)
	assert.NotNil(t,checktemplate)
	assert.Nil(t,err)

	// assert equality
	assert.Equal(t, 123, 123, "they should be equal")

	// assert inequality
	assert.NotEqual(t, 123, 456, "they should not be equal")

	// assert for nil (good for errors)
	assert.Nil(t, nil)

	// assert for not nil (good when you expect something)
	if assert.NotNil(t, 2) {

		// now we know that object isn't nil, we are safe to make
		// further assertions without causing any errors
		assert.Equal(t, "Something", "Something")

	}

}
