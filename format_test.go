package librarianpuppetgo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormat(t *testing.T) {
	mods, _ := parsePuppetfile(r(`
mod 'foo',:git=>'aaabbb',:ref=>'fix/a-bug'
mod 'bar',:git=>'cccddd',:ref=>'dev'`))
	s := format(mods)
	assert.Equal(t, `mod 'bar', :git => 'cccddd', :ref => 'dev'
mod 'foo', :git => 'aaabbb', :ref => 'fix/a-bug'
`, s)
}
