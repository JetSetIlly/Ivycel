package cells_test

import (
	"errors"
	"testing"

	"github.com/jetsetilly/ivycel/cells"
)

func ExpectEquality[T comparable](t *testing.T, value T, expectedValue T) {
	t.Helper()
	if value != expectedValue {
		t.Errorf("equality test of type %T failed: '%v' does not equal '%v')", value, value, expectedValue)
	}
}

func ExpectedError(t *testing.T, err error, expected error) {
	t.Helper()
	if !errors.Is(err, expected) {
		t.Errorf("%v is an unexpected error", err)
	}
}

func TestPositionReference(t *testing.T) {
	var p cells.Position

	ExpectEquality(t, p.String(), "A1")
	p.Column = 1
	ExpectEquality(t, p.String(), "B1")
	p.Column = 2
	ExpectEquality(t, p.String(), "C1")
	p.Column = 23
	ExpectEquality(t, p.String(), "X1")
	p.Column = 24
	ExpectEquality(t, p.String(), "Y1")
	p.Column = 25
	ExpectEquality(t, p.String(), "Z1")

	p.Column = 26
	ExpectEquality(t, p.String(), "AA1")
	p.Column = 52
	ExpectEquality(t, p.String(), "BA1")
}

func TestPositionFromReference(t *testing.T) {
	type testSpec struct {
		ref        string
		err        error
		normalised string
	}
	tests := []testSpec{
		{ref: "A1", err: nil},
		{ref: "B1", err: nil},
		{ref: "A99", err: nil},
		{ref: "Z99", err: nil},
		{ref: "AA99", err: nil},
		{ref: "AB12", err: nil},
		{ref: "BA498", err: nil},
		{ref: "QUX100", err: nil},

		// illegal references
		{ref: "A0", err: cells.IllegalReference},
		{ref: "00", err: cells.IllegalReference},
		{ref: "ZZ", err: cells.IllegalReference},
		{ref: "A0Z", err: cells.IllegalReference},

		// some conversions will result in a normalised form
		{ref: "A01", err: nil, normalised: "A1"},
		{ref: "a1", err: nil, normalised: "A1"},
		{ref: "bA498", err: nil, normalised: "BA498"},
		{ref: " A1", err: nil, normalised: "A1"},
		{ref: "A99 ", err: nil, normalised: "A99"},
	}

	for _, tst := range tests {
		p, err := cells.PositionFromReference(tst.ref)
		ExpectedError(t, err, tst.err)
		if err == nil {
			if tst.normalised == "" {
				ExpectEquality(t, p.String(), tst.ref)
			} else {
				ExpectEquality(t, p.String(), tst.normalised)
			}
		}
	}
}
