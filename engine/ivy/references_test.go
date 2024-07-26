package ivy_test

import (
	"fmt"
	"testing"

	"github.com/jetsetilly/ivycel/engine/ivy"
)

func ExpectEquality[T comparable](t *testing.T, value T, expectedValue T) {
	t.Helper()
	if value != expectedValue {
		t.Errorf("equality test of type %T failed: '%v' does not equal '%v')", value, value, expectedValue)
	}
}

func TestCellReference(t *testing.T) {
	var ok bool

	// cell references must be wrapped
	ok = ivy.CellReferenceMatch.MatchString("A1")
	ExpectEquality(t, ok, false)

	ok = ivy.CellReferenceMatch.MatchString("ZZZ999")
	ExpectEquality(t, ok, false)

	// correctly wrapped references
	ok = ivy.CellReferenceMatch.MatchString("{A1}")
	ExpectEquality(t, ok, true)

	ok = ivy.CellReferenceMatch.MatchString("{AA100}")
	ExpectEquality(t, ok, true)

	ok = ivy.CellReferenceMatch.MatchString("{ZXC234}")
	ExpectEquality(t, ok, true)

	// the order of letters and digits is important
	ok = ivy.CellReferenceMatch.MatchString("{1A}")
	ExpectEquality(t, ok, false)

	ok = ivy.CellReferenceMatch.MatchString("{100AA}")
	ExpectEquality(t, ok, false)

	ok = ivy.CellReferenceMatch.MatchString("{1A}")
	ExpectEquality(t, ok, false)

	// leading/trailing space is allowed
	ok = ivy.CellReferenceMatch.MatchString(" {A1}")
	ExpectEquality(t, ok, true)

	ok = ivy.CellReferenceMatch.MatchString("{A1} ")
	ExpectEquality(t, ok, true)

	// internal spaces not allowed
	ok = ivy.CellReferenceMatch.MatchString("{ A1}")
	ExpectEquality(t, ok, false)

	ok = ivy.CellReferenceMatch.MatchString("{A 1}")
	ExpectEquality(t, ok, false)

	ok = ivy.CellReferenceMatch.MatchString("{A1 }")
	ExpectEquality(t, ok, false)

	// multiple instances can appear in the string
	ok = ivy.CellReferenceMatch.MatchString(" {A1} + {B100}")
	ExpectEquality(t, ok, true)

	ok = ivy.CellReferenceMatch.MatchString(" {A1} + {B100} / {ZZZ2309}")
	ExpectEquality(t, ok, true)
}

// these strings should match after wrapping. wrapping should be part of the test loop
var expectedSuccess = []string{
	"A1",
	"ZZ100",
}

// these strings shouldn't match, even after wrapping
var expectedFail = []string{
	" A1",
	"ZZ 100",
	"1A",
	"100ZZ",
}

func TestCellReferenceReplace(t *testing.T) {
	// the prefix we use for replacement isn't important
	prefix := "v"

	for _, s := range expectedSuccess {
		w := ivy.WrapCellReference(s)
		r := ivy.CellReferenceMatch.ReplaceAllString(w, fmt.Sprintf("%s$1", prefix))
		ExpectEquality(t, r, fmt.Sprintf("%s%s", prefix, s))
	}

	for _, s := range expectedFail {
		w := ivy.WrapCellReference(s)
		r := ivy.CellReferenceMatch.ReplaceAllString(w, fmt.Sprintf("%s$1", prefix))
		ExpectEquality(t, r, w)
	}
}

func TestCellReferenceSubmatch(t *testing.T) {
	var matches [][]string

	for _, s := range expectedSuccess {
		w := ivy.WrapCellReference(s)
		matches = ivy.CellReferenceMatch.FindAllStringSubmatch(w, 1)
		ExpectEquality(t, len(matches), 1)
		ExpectEquality(t, len(matches[0]), 2)
		ExpectEquality(t, matches[0][1], s)
	}

	for _, s := range expectedFail {
		w := ivy.WrapCellReference(s)
		matches = ivy.CellReferenceMatch.FindAllStringSubmatch(w, 1)
		ExpectEquality(t, len(matches), 0)
	}
}
