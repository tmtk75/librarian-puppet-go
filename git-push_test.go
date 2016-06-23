package librarianpuppetgo

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func newGit() *Git {
	return &Git{
		Remote: "origin",
		IsCommit: func(wd, sha1 string) bool {
			if sha1 == "a-sha1-abcd1234" {
				return true
			}
			return false
		},
		IsBranch: func(wd, name string) bool {
			if name == "master" || name == "a-topic" || name == "release/0.2" {
				return true
			}
			return false
		},
		IsTag: func(wd, name string) bool {
			if name == "v0.1.3" || name == "v0.2.3" {
				return true
			}
			return false
		},
		Sha1: func(wd, ref string) string {
			if ref == "a-topic" {
				return "a-topic-sha1"
			}
			return ""
		},
		Diff: func(wd, srcref, dstref string) string {
			return "a"
		},
	}

}

func TestGitPushCmd(t *testing.T) {
	var tests = []struct {
		sha1nil bool
		errnil  bool
		src     string
		dst     string
		exp     string
	}{
		{false, true, "release/0.1", "a-sha1-abcd1234", "(cd modules/foo; git branch a-sha1-abcd1234 release/0.2; git push origin release/0.2:release/0.2)"},
		{false, true, "release/0.1", "a-topic", "(cd modules/foo; git branch a-topic-sha1 release/0.2; git push origin release/0.2:release/0.2)"},
		{false, true, "not-a-branch", "a-sha1-abcd1234", "# INFO: modules/foo is referred at a-sha1-abcd1234"},
		{true, true, "release/0.1", "master", "(cd modules/foo; git push origin master:release/0.2)"},
		{false, false, "release/0.1", "develop", ""},
	}
	for _, c := range tests {
		git := newGit()
		if c.sha1nil {
			git.Sha1 = nil
		}
		s, err := git.PushCmd(
			Mod{name: "foo", opts: map[string]string{"ref": c.src}},
			Mod{name: "foo", opts: map[string]string{"ref": c.dst}},
		)
		if c.errnil {
			assert.Nil(t, err)
		} else {
			assert.NotNil(t, err)
		}
		assert.Equal(t, c.exp, s)
	}
}

func TestGitPushCmd2(t *testing.T) {
	git := newGit()
	//
	a, _ := parsePuppetfile(r(`mod 'puppetlabs/foo', '3.0.3'`))
	b, _ := parsePuppetfile(r(`mod 'puppetlabs/foo', '3.0.3'`))
	s, err := git.PushCmd(a[0], b[0])
	assert.Nil(t, err)
	assert.Equal(t, "# puppetlabs/foo 3.0.3 doesn't have :ref", s)

	//
	a, _ = parsePuppetfile(r(`mod 'puppetlabs/bar'`))
	b, _ = parsePuppetfile(r(`mod 'puppetlabs/bar'`))
	s, err = git.PushCmd(a[0], b[0])
	assert.Nil(t, err)
	assert.Equal(t, "# puppetlabs/bar doesn't have :ref", s)
}

func TestGitPushCmdTag(t *testing.T) {
	var tests = []struct {
		diff string
		src  string
		dst  string
		exp  string
	}{
		{"", "v0.1.3", "release/0.2", "# INFO: no diff for foo between v0.1.3 and release/0.2"},
		{"a", "v0.1.3", "release/0.2", "(cd modules/foo; git tag v0.2.0 release/0.2; git push origin v0.2.0)"},
		{"a", "v0.1.3", "a-topic", "# WARN: a-topic cannot be parsed minor version for foo"},
		{"", "v0.2.3", "release/0.2", "# INFO: no diff for foo between v0.2.3 and release/0.2"},
		{"a", "v0.2.3", "release/0.2", "(cd modules/foo; git tag v0.2.4 release/0.2; git push origin v0.2.4)"},
	}
	//

	for _, c := range tests {
		git := newGit()
		git.Diff = func(wd, srcref, dstref string) string { return c.diff }
		s, err := git.PushCmd(
			Mod{name: "foo", user: "bar", opts: map[string]string{"ref": c.src}},
			Mod{name: "foo", user: "bar", opts: map[string]string{"ref": c.dst}},
		)
		assert.Nil(t, err)
		assert.Equal(t, c.exp, s)
	}
}

func TestMinorVersionNumber(t *testing.T) {
	v, err := minorVersionNumber("master")
	assert.NotNil(t, err)
	assert.Equal(t, -1, v)

	v, err = minorVersionNumber("release/0.x")
	assert.NotNil(t, err)
	assert.Equal(t, -1, v)

	v, err = minorVersionNumber("release/0.2")
	assert.Nil(t, err)
	assert.Equal(t, 2, v)
}

func TestSemanticVersion(t *testing.T) {
	ma, mi, tri, _ := semanticVersion("v10.200.3000")
	assert.Equal(t, 10, ma)
	assert.Equal(t, 200, mi)
	assert.Equal(t, 3000, tri)

	ma, mi, tri, _ = semanticVersion("10.200.3000")
	assert.Equal(t, 10, ma)
	assert.Equal(t, 200, mi)
	assert.Equal(t, 3000, tri)

	_, _, _, err := semanticVersion("x.y.z")
	assert.NotNil(t, err)

	x, y, z, err := semanticVersion("1.2")
	assert.Nil(t, err)
	assert.Equal(t, 1, x)
	assert.Equal(t, 2, y)
	assert.Equal(t, -1, z)

	x, y, z, err = semanticVersion("v3.10")
	assert.Nil(t, err)
	assert.Equal(t, 3, x)
	assert.Equal(t, 10, y)
	assert.Equal(t, -1, z)
}

func TestGitPushCmds(t *testing.T) {
	git := newGit()

	//
	buf := bytes.NewBuffer([]byte{})
	git.Writer = buf
	git.PushCmds("./files/git-push.src", "./files/empty")
	assert.Equal(t, "# foo is missing in ./files/empty\n", buf.String())

	//
	buf = bytes.NewBuffer([]byte{})
	git.Writer = buf
	git.PushCmds("./files/git-push-cmds.src", "./files/git-push-cmds.dst")
	assert.Equal(t, strings.TrimSpace(`
(cd modules/foo; git branch a-sha1-abcd1234 release/0.2; git push origin release/0.2:release/0.2)
(cd modules/bar; git branch a-topic-sha1 release/0.2; git push origin release/0.2:release/0.2)
# INFO: modules/fiz is referred at a-sha1-abcd1234
	`)+"\n", buf.String())
}
