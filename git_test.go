package librarianpuppetgo

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsCommit(t *testing.T) {
	assert.True(t, isCommit(".", "f8d13558bafc452e6994c015ac807e367e0fb557"))
	assert.False(t, isCommit(".", "v0.1.0"))
	assert.False(t, isCommit(".", "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"))
}

func TestIsTag(t *testing.T) {
	assert.True(t, isTag(".", "v0.1.0"))
	assert.False(t, isTag(".", "no-tag"))
}

func TestIsBranch(t *testing.T) {
	assert.True(t, isBranch(".", "master"))
	assert.False(t, isBranch(".", "no-branch"))
}

func TestGitSha1(t *testing.T) {
	assert.Equal(t, "fb7715971404342c765d05524fe50bcdb982a5e8", gitSha1(".", "v0.1.0"))
}

func TestRunTimeout(t *testing.T) {
	timeout = 2
	err := run(".", "sleep", []string{"10"}) // timed out in 2 seconds
	if err == nil {
		t.Errorf("should be error")
	}
	if err != context.DeadlineExceeded {
		t.Errorf("should be error %v", err)
	}
}
