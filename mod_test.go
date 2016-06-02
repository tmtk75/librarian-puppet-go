package librarianpuppetgo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestModFormat(t *testing.T) {
	mods, _ := parsePuppetfile(r(`mod 'foo', :git => 'user@github.com/foo/bar', :ref => 'fix/a-bug'`))
	assert.Equal(t, "mod 'foo', :git => 'user@github.com/foo/bar', :ref => 'fix/a-bug'", mods[0].Format())

	mods, _ = parsePuppetfile(r(`mod 'foo/bar', :git => 'a@b.com', :ref => '1.0.0'`))
	assert.Equal(t, "mod 'foo/bar', :git => 'a@b.com', :ref => '1.0.0'", mods[0].Format())

	mods, _ = parsePuppetfile(r(`mod 'puppetlabs/stdlib', '4.1.0'`))
	assert.Equal(t, "mod 'puppetlabs/stdlib', '4.1.0'", mods[0].Format())

	mods, _ = parsePuppetfile(r(`mod 'foobar/brabra'`))
	assert.Equal(t, "mod 'foobar/brabra'", mods[0].Format())
}
