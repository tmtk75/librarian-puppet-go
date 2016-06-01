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

func TestBumpUpMod(t *testing.T) {
	nodiff := func(wd, a, b string) string { return "" }
	diff := func(wd, a, b string) string { return "a" }

	a, _ := parsePuppetfile(r(``))
	b, _ := parsePuppetfile(r(`mod 'dprince/qpid'`))
	e := bumpUpMod(b[0], a, "foo", "a-file", nodiff)
	assert.Equal(t, "mod 'dprince/qpid'", e)

	a, _ = parsePuppetfile(r(``))
	b, _ = parsePuppetfile(r(`mod 'fiz/buz', :git => 'abc', :ref => '01234'`))
	e = bumpUpMod(b[0], a, "initial", "a-file", nodiff)
	assert.Equal(t, "mod 'fiz/buz', :git => 'abc', :ref => 'initial'", e)

	a, _ = parsePuppetfile(r(`mod 'a/b', :git => 'aaa'`))
	b, _ = parsePuppetfile(r(`mod 'a/b', :git => 'aaa', :ref => 'development'`))
	e = bumpUpMod(b[0], a, "foo", "a-file", diff)
	assert.Equal(t, "mod 'a/b', :git => 'aaa', :ref => 'development'", e)

	a, _ = parsePuppetfile(r(`mod 'a/b', :git => 'aaa', :ref => 'release/0.1'`))
	b, _ = parsePuppetfile(r(`mod 'a/b', :git => 'aaa', :ref => 'development'`))
	e = bumpUpMod(b[0], a, "foo", "a-file", nodiff)
	assert.Equal(t, "mod 'a/b', :git => 'aaa', :ref => 'release/0.1'", e)

	a, _ = parsePuppetfile(r(`mod 'a/b', :git => 'aaa', :ref => '0123456789a'`))
	b, _ = parsePuppetfile(r(`mod 'a/b', :git => 'aaa', :ref => 'development'`))
	e = bumpUpMod(b[0], a, "foo", "a-file", nodiff)
	assert.Equal(t, "mod 'a/b', :git => 'aaa', :ref => '0123456789a'", e)

	a, _ = parsePuppetfile(r(`mod 'a/b', :git => 'aaa', :ref => 'release/0.1'`))
	b, _ = parsePuppetfile(r(`mod 'a/b', :git => 'aaa', :ref => 'development'`))
	e = bumpUpMod(b[0], a, "foo", "a-file", diff)
	assert.Equal(t, "mod 'a/b', :git => 'aaa', :ref => 'release/0.2'", e)
}
