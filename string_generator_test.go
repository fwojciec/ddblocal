package ddblocal_test

import (
	"testing"

	"github.com/fwojciec/ddblocal"
)

func TestGeneratesStringsOfCorrectLength(t *testing.T) {
	t.Parallel()
	sg := ddblocal.NewStringGenerator()
	res, err := sg.Generate()
	ok(t, err)
	equals(t, 12, len(res))
}

func TestGeneratesUniqueStrings(t *testing.T) {
	t.Parallel()
	sg := ddblocal.NewStringGenerator()
	res1, err := sg.Generate()
	ok(t, err)
	res2, err := sg.Generate()
	ok(t, err)
	assert(t, res1 != res2, "expected generated strings to be different")
}
