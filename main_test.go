package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPatternCIIC(t *testing.T) {
	dna = DNA("CIIC")
	p, err := pattern()
	assert.NoError(t, err)
	assert.Len(t, p, 1)
	assert.ElementsMatch(t, p, []interface{}{'I'})
}

func TestPatternIIPIPICPIICICIIF(t *testing.T) {
	dna = DNA("IIPIPICPIICICIIF")
	p, err := pattern()
	assert.NoError(t, err)
	assert.Len(t, p, 4)
	assert.ElementsMatch(t, p, []interface{}{true, 2, false, 'P'})
}

func TestConstsCFICPII(t *testing.T) {
	dna = DNA("CFICPII")
	c := consts()
	assert.Len(t, c, 4)
	assert.Equal(t, c, "ICPF")
	assert.Equal(t, dna, DNA("II"))
}

func TestFindPostfix(t *testing.T) {
	n := findPostfix(DNA("ICFICFICPF"), "ICP")
	assert.Equal(t, 9, n)

	n = findPostfix(DNA("ICFICFICPF"), "PF")
	assert.Equal(t, 10, n)

	n = findPostfix(DNA("ICFICFICPF"), "PFI")
	assert.Equal(t, -1, n)

	n = findPostfix(DNA("ICFICFICPF"), "I")
	assert.Equal(t, 1, n)
}

func TestAsNAt(t *testing.T) {
	// 4 => IICP
	// 0 => P
	// 5 => CICP

	s := asnat(4)
	assert.Equal(t, "IICP", string(s))

	s = asnat(0)
	assert.Equal(t, "P", string(s))

	s = asnat(5)
	assert.Equal(t, "CICP", string(s))
}
