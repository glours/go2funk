package collection

import (
	"fmt"
	"gotest.tools/v3/assert"
	"testing"
)

var (
	entry = NewEntry[int, string](10, "ten", nil)
)

func TestGetKey(t *testing.T) {
	assert.Equal(t, entry.GetKey(), 10, fmt.Sprintf("key should be 10 but is %d", entry.GetKey()))
}

func TestGetValue(t *testing.T) {
	assert.Equal(t, entry.GetValue(), "ten", fmt.Sprintf("value should be 'ten' but is '%s'", entry.GetValue()))
}

func TestEquals(t *testing.T) {
	assert.Assert(t, entry.Equals(NewEntry[int, string](10, "ten", nil)))
}

func TestHashCode(t *testing.T) {
	var hashCheck uint64 = 16887331970960175425
	hash, err := entry.HashCode()
	assert.NilError(t, err)
	assert.Equal(t, hash, hashCheck, fmt.Sprintf("Hash value should be %d but is %d", hashCheck, hash))
}
