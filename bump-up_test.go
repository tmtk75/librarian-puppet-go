package librarianpuppetgo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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
