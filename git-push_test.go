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
			if name == "v0.1.3" {
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
	git := newGit()
	//
	s, err := git.PushCmd(
		Mod{name: "foo", opts: map[string]string{"ref": "release/0.1"}},
		Mod{name: "foo", opts: map[string]string{"ref": "a-sha1-abcd1234"}},
	)
	assert.Nil(t, err)
	assert.Equal(t, "(cd modules/foo; git branch a-sha1-abcd1234 release/0.2; git push origin release/0.2:release/0.2)", s)

	//
	s, err = git.PushCmd(
		Mod{name: "foo", opts: map[string]string{"ref": "release/0.1"}},
		Mod{name: "foo", opts: map[string]string{"ref": "a-topic"}},
	)
	assert.Nil(t, err)
	assert.Equal(t, "(cd modules/foo; git branch a-topic-sha1 release/0.2; git push origin release/0.2:release/0.2)", s)

	//
	s, err = git.PushCmd(
		Mod{name: "foo", opts: map[string]string{"ref": "not-a-branch"}},
		Mod{name: "foo", opts: map[string]string{"ref": "a-sha1-abcd1234"}},
	)
	assert.Nil(t, err)
	assert.Equal(t, "# modules/foo is referred at a-sha1-abcd1234", s)

	//
	git.Sha1 = nil
	s, err = git.PushCmd(
		Mod{name: "foo", opts: map[string]string{"ref": "release/0.1"}},
		Mod{name: "foo", opts: map[string]string{"ref": "master"}},
	)
	assert.Nil(t, err)
	assert.Equal(t, "(cd modules/foo; git push origin master:release/0.2)", s)

	//
	s, err = git.PushCmd(
		Mod{name: "foo", opts: map[string]string{"ref": "release/0.1"}},
		Mod{name: "foo", opts: map[string]string{"ref": "develop"}},
	)
	assert.NotNil(t, err)
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
	git := newGit()
	//
	s, err := git.PushCmd(
		Mod{name: "foo", user: "bar", opts: map[string]string{"ref": "v0.1.3"}},
		Mod{name: "foo", user: "bar", opts: map[string]string{"ref": "release/0.2"}},
	)
	assert.Nil(t, err)
	assert.Equal(t, "(cd modules/foo; git push origin release/0.2:v0.2.0)", s)

	//
	s, err = git.PushCmd(
		Mod{name: "foo", user: "bar", opts: map[string]string{"ref": "v0.1.3"}},
		Mod{name: "foo", user: "bar", opts: map[string]string{"ref": "a-topic"}},
	)
	assert.Nil(t, err)
	assert.Equal(t, "# WARN: a-topic cannot be parsed minor version for foo", s)
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
# modules/fiz is referred at a-sha1-abcd1234
	`)+"\n", buf.String())
}
