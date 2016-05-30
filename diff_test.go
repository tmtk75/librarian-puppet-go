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
