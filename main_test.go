package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPatternCIIC(t *testing.T) {
	dna = dnaFromString("CIIC")
	p, err := pattern(dna.iterator())
	assert.NoError(t, err)
	assert.Len(t, p, 1)
	assert.ElementsMatch(t, p, []interface{}{'I'})
}

func TestPatternIIPIPICPIICICIIF(t *testing.T) {
	dna = dnaFromString("IIPIPICPIICICIIF")
	p, err := pattern(dna.iterator())
	assert.NoError(t, err)
	assert.Len(t, p, 4)
	assert.ElementsMatch(t, p, []interface{}{true, 2, false, 'P'})
}

func TestConstsCFICPII(t *testing.T) {
	dna = dnaFromString("CFICPII")
	iter := dna.iterator()
	c := consts(iter)
	assert.Len(t, c, 4)
	assert.Equal(t, c, "ICPF")
	assert.Equal(t, dnaFromString("II"), iter.Rest())
}

func TestFindPostfix(t *testing.T) {
	n := findPostfix(dnaFromString("ICFICFICPF"), "ICP")
	assert.Equal(t, 9, n)

	n = findPostfix(dnaFromString("ICFICFICPF"), "PF")
	assert.Equal(t, 10, n)

	n = findPostfix(dnaFromString("ICFICFICPF"), "PFI")
	assert.Equal(t, -1, n)

	n = findPostfix(dnaFromString("ICFICFICPF"), "I")
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

func TestRun(t *testing.T) {
	err := do("")
	assert.NoError(t, err)
}

func TestIterator(t *testing.T) {
	d := &DNA{
		s:   "",
		len: 5,
		left: &DNA{
			s:     "",
			len:   2,
			left:  &DNA{s: "a", len: 1},
			right: &DNA{s: "b", len: 1},
		},
		right: &DNA{
			s:     "",
			len:   3,
			left:  &DNA{s: "c", len: 1},
			right: &DNA{s: "de", len: 2},
		},
	}

	iter := d.iterator()
	assert.Equal(t, byte('a'), iter.Next())
	assert.Equal(t, "bcde", iter.Rest().asString())
	assert.Equal(t, byte('b'), iter.Next())
	assert.Equal(t, "cde", iter.Rest().asString())
	assert.Equal(t, byte('c'), iter.Next())
	assert.Equal(t, "de", iter.Rest().asString())
	assert.Equal(t, byte('d'), iter.Next())
	assert.Equal(t, "e", iter.Rest().asString())
	assert.Equal(t, byte('e'), iter.Next())
	assert.Equal(t, "", iter.Rest().asString())
	assert.Equal(t, byte(0), iter.Next())
}
