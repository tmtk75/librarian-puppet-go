package librarianpuppetgo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBumpUpMod(t *testing.T) {
	var nodiff = func(wd, a, b string) string { return "" }
	var diff = func(wd, a, b string) string { return "a" }
	isBranch := func(wd, name string) bool {
		if name == "release/0.2" || name == "release/foobar" {
			return true
		}
		return false
	}
	isTag := func(wd, name string) bool {
		if name == "v0.1.3" || name == "v0.2.1" {
			return true
		}
		return false
	}

	tests := []struct {
		rel  string
		diff bumpDiff
		err  bool
		src  string
		dst  string
		exp  string
	}{
		//
		{
			"", nodiff, false,
			``,
			`mod 'dprince/qpid'`,
			`mod 'dprince/qpid'`},
		{
			"initial", nodiff, false,
			``,
			`mod 'fiz/buz', :git => 'abc', :ref => '01234'`,
			`mod 'fiz/buz', :git => 'abc', :ref => 'initial'`},
		{
			"", diff, false,
			`mod 'a/b', :git => 'aaa'`,
			`mod 'a/b', :git => 'aaa', :ref => 'development'`,
			`mod 'a/b', :git => 'aaa', :ref => 'development'`},
		{
			"", nodiff, false,
			`mod 'a/b', :git => 'aaa', :ref => 'release/0.1'`,
			`mod 'a/b', :git => 'aaa', :ref => 'development'`,
			`mod 'a/b', :git => 'aaa', :ref => 'release/0.1'`},
		{
			"", nodiff, false,
			`mod 'a/b', :git => 'aaa', :ref => '0123456789a'`,
			`mod 'a/b', :git => 'aaa', :ref => 'development'`,
			`mod 'a/b', :git => 'aaa', :ref => '0123456789a'`},
		{
			"", diff, false,
			`mod 'a/b', :git => 'aaa', :ref => 'release/0.1'`,
			`mod 'a/b', :git => 'aaa', :ref => 'development'`,
			`mod 'a/b', :git => 'aaa', :ref => 'release/0.2'`},
		//
		{
			"", diff, false,
			`mod 'a/b', :git => 'aaa', :ref => 'v0.1.3'`,
			`mod 'a/b', :git => 'aaa', :ref => 'release/0.2'`,
			`mod 'a/b', :git => 'aaa', :ref => 'v0.2.0'`},
		{
			"", nodiff, false,
			`mod 'a/b', :git => 'aaa', :ref => 'v0.2.1'`,
			`mod 'a/b', :git => 'aaa', :ref => 'release/0.2'`,
			`mod 'a/b', :git => 'aaa', :ref => 'v0.2.1'`},
		{
			"", diff, false,
			`mod 'a/b', :git => 'aaa', :ref => 'v0.2.1'`,
			`mod 'a/b', :git => 'aaa', :ref => 'release/0.2'`,
			`mod 'a/b', :git => 'aaa', :ref => 'v0.2.2'`},
		{
			"", diff, true,
			`mod 'a/b', :git => 'aaa', :ref => 'v0.2.1'`,
			`mod 'a/b', :git => 'aaa', :ref => 'release/foobar'`,
			``},
	}

	for _, c := range tests {
		a, _ := parsePuppetfile(r(c.src))
		b, _ := parsePuppetfile(r(c.dst))
		g := NewGit()
		g.Diff = c.diff
		g.IsTag = isTag
		g.IsBranch = isBranch
		e, err := g.bumpUpMod(b[0], a, c.rel, "a-file")
		if c.err {
			assert.NotNil(t, err)
		} else {
			assert.Nil(t, err)
		}
		assert.Equal(t, c.exp, e)
	}
}
