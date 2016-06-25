package librarianpuppetgo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReplaceWithMod(t *testing.T) {
	tests := []struct {
		templ string
		mod   Mod
		exp   string
		err   bool
	}{
		{
			"{{.Name}} {{.Ref}}",
			Mod{name: "foo", opts: map[string]string{"ref": "01234"}},
			"foo 01234", false,
		},
		{
			"a\\nb\\t",
			Mod{},
			"a\nb\t", false,
		},
	}

	for _, c := range tests {
		r, err := replaceWithMod(c.templ, c.mod)
		if c.err {
			assert.NotNil(t, err)
		} else {
			assert.Nil(t, err)
		}
		assert.Equal(t, c.exp, r)
	}
}

func TestMakeEachArgs(t *testing.T) {
	mod := Mod{name: "foo", opts: map[string]string{"ref": "01234"}}

	r, err := makeEachArgs([]string{"git", "log", "{{.Name}}", "{{.Ref}}"}, mod)
	assert.Nil(t, err)
	assert.Equal(t, []string{"git", "log", "foo", "01234"}, r)

	r, err = makeEachArgs([]string{"{{.Name}"}, mod)
	assert.NotNil(t, err)
	assert.Equal(t, []string{"{{.Name}"}, r)
}
