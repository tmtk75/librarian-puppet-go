package librarianpuppetgo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSort(t *testing.T) {
	//
	a, err := Sort(`
	0.1
	0.2
        
	0.10`)
	assert.Nil(t, err)
	assert.Equal(t, `0.1
0.2
0.10`, a)

	//
	a, err = Sort(`
	0.11
	0.2  	
        
	0.10`)
	assert.Nil(t, err)
	assert.Equal(t, `0.2
0.10
0.11`, a)

	//
	a, err = Sort(`
	v0.11.1
	v0.2.10
        
	v0.10.12`)
	assert.Nil(t, err)
	assert.Equal(t, `0.2.10
0.10.12
0.11.1`, a)

	//
	a, err = Sort(`
	0.11
	0.2
        
	0.11.1`)
	assert.Nil(t, err)
	assert.Equal(t, `0.2
0.11
0.11.1`, a)

	//
	a, err = Sort(`
	v0.11
	0.2
        
	v0.11.1`)
	assert.Nil(t, err)
	assert.Equal(t, `0.2
0.11
0.11.1`, a)

}
