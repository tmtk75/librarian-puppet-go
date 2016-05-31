package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIncrement(t *testing.T) {
	a, _ := increment("release/0.3")
	assert.Equal(t, "release/0.4", a)

	a, _ = increment("release/0.9")
	assert.Equal(t, "release/0.10", a)

	a, _ = increment("release/0.10")
	assert.Equal(t, "release/0.11", a)

	_, err := increment("rel/0.10")
	assert.NotNil(t, err)

	_, err = increment("release/0.x")
	assert.NotNil(t, err)
}

func TestModString(t *testing.T) {
	mods, _ := parsePuppetfile(r(`mod 'foo', :git => 'user@github.com/foo/bar', :ref => 'fix/a-bug'`))
	assert.Equal(t, "mod 'foo', :git => 'user@github.com/foo/bar', :ref => 'fix/a-bug'", mods[0].Format())

	mods, _ = parsePuppetfile(r(`mod 'foo/bar', :git => 'a@b.com', :ref => '1.0.0'`))
	assert.Equal(t, "mod 'foo/bar', :git => 'a@b.com', :ref => '1.0.0'", mods[0].Format())

	mods, _ = parsePuppetfile(r(`mod 'puppetlabs/stdlib', '4.1.0'`))
	assert.Equal(t, "mod 'puppetlabs/stdlib', '4.1.0'", mods[0].Format())

	mods, _ = parsePuppetfile(r(`mod 'foobar/brabra'`))
	assert.Equal(t, "mod 'foobar/brabra'", mods[0].Format())
}
