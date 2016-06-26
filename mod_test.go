package librarianpuppetgo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestModFullname(t *testing.T) {
	mods, _ := parsePuppetfile(r(`mod 'foo', :ref => '0.1.0'`))
	assert.Equal(t, "foo", mods[0].Fullname())

	mods, _ = parsePuppetfile(r(`mod 'bar/foo', '0.1.0'`))
	assert.Equal(t, "bar/foo", mods[0].Fullname())
}

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

func TestModRef(t *testing.T) {
	mods, _ := parsePuppetfile(r(`mod 'foo', :git => 'user@github.com/foo/bar', :ref => 'fix/a-bug'`))
	assert.Equal(t, "fix/a-bug", mods[0].Ref())

	mods, _ = parsePuppetfile(r(`mod 'foo/bar', :git => 'a@b.com', :ref => '1.0.0'`))
	assert.Equal(t, "1.0.0", mods[0].Ref())

	mods, _ = parsePuppetfile(r(`mod 'puppetlabs/stdlib', '4.1.0'`))
	assert.Equal(t, "4.1.0", mods[0].Ref())

	mods, _ = parsePuppetfile(r(`mod 'foobar/brabra'`))
	assert.Equal(t, "", mods[0].Ref())
}
