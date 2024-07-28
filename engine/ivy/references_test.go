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

}

func TestCellReferenceWithIndexing(t *testing.T) {
	var ok bool

	// indexing is not explicitely allowed but should be captured and used in
	// the ivy expression as far as possible

	ok = ivy.CellReferenceMatch.MatchString("{A1[0]}")
	ExpectEquality(t, ok, true)
	ok = ivy.CellReferenceMatch.MatchString("{A1[0][1]}")
	ExpectEquality(t, ok, true)

	// this is an illegal reference as far as ivy is concerned but we don't
	// worry about that and pass it on as normal
	ok = ivy.CellReferenceMatch.MatchString("{A1[}")
	ExpectEquality(t, ok, true)
	ok = ivy.CellReferenceMatch.MatchString("{A1[a[[]}")
	ExpectEquality(t, ok, true)

	// but spaces are detected and not allowed
	ok = ivy.CellReferenceMatch.MatchString("{A1 [}")
	ExpectEquality(t, ok, false)
	ok = ivy.CellReferenceMatch.MatchString("{A1[ 0]}")
	ExpectEquality(t, ok, false)
	ok = ivy.CellReferenceMatch.MatchString("{A1[0] [ 1]}")
	ExpectEquality(t, ok, false)
}

// the prefix we use for replacement isn't important
const prefix = "v"

func TestCellReferenceByTable(t *testing.T) {
	type test struct {
		inp   string
		match string
		conv  string
	}

	var testingTable = []test{
		{inp: "A1", match: "A1", conv: fmt.Sprintf("%sA1", prefix)},
		{inp: "ZZ100", match: "ZZ100", conv: fmt.Sprintf("%sZZ100", prefix)},
		{inp: " A1"},
		{inp: "ZZ 100"},
		{inp: "1A"},
		{inp: "100ZZ"},

		// indexing tests
		{inp: "A1[0]", match: "A1[0]", conv: fmt.Sprintf("%sA1[0]", prefix)},

		// even though we know for sure that [0 is not a valid index we need to
		// identify it and pass it to ivy for parsing
		{inp: "A1[0", match: "A1[0", conv: fmt.Sprintf("%sA1[0", prefix)},
	}

	var matches [][]string

	for _, tst := range testingTable {
		w := ivy.WrapCellReference(tst.inp)
		matches = ivy.CellReferenceMatch.FindAllStringSubmatch(w, 1)
		if tst.match != "" {
			ExpectEquality(t, len(matches), 1)
			ExpectEquality(t, len(matches[0]), 2)
			ExpectEquality(t, matches[0][1], tst.match)

			r := ivy.CellReferenceMatch.ReplaceAllString(w, fmt.Sprintf("%s$1", prefix))
			ExpectEquality(t, r, tst.conv)
		} else {
			ExpectEquality(t, len(matches), 0)

			// replacement will fail so returned string should equal the wrapped string
			r := ivy.CellReferenceMatch.ReplaceAllString(w, fmt.Sprintf("%s$1", prefix))
			ExpectEquality(t, r, w)
		}
	}
}

func TestCellReferenceInExpressions(t *testing.T) {
	type test struct {
		inp  string
		repl string
	}

	var testingTable []test = []test{
		{
			inp:  "{A1} + {B100}",
			repl: "vA1 + vB100",
		},
		{
			inp:  "{A1} + {B100} / {ZZZ2309}",
			repl: "vA1 + vB100 / vZZZ2309",
		},
		{
			inp:  "{A1} + {B100[1]} / {ZZZ2309[100][203]}",
			repl: "vA1 + vB100[1] / vZZZ2309[100][203]",
		},
	}

	for _, tst := range testingTable {
		r := ivy.CellReferenceMatch.ReplaceAllString(tst.inp, fmt.Sprintf("%s$1", prefix))
		ExpectEquality(t, r, tst.repl)
	}
}
