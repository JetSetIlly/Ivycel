package references_test

import (
	"fmt"
	"testing"

	"github.com/jetsetilly/ivycel/cells"
	"github.com/jetsetilly/ivycel/references"
)

func TestExpressions(t *testing.T) {
	type test struct {
		from string
		to   string
	}

	var testingTable []test = []test{
		{
			from: "{A1} + {B100}",
			to:   "vA1 + vB100",
		},
		{
			from: "{A1} + {B100} / {ZZZ2309}",
			to:   "vA1 + vB100 / vZZZ2309",
		},
		{
			from: "{A1} + {B100[1]} / {ZZZ2309[100][203]}",
			to:   "vA1 + vB100[1] / vZZZ2309[100][203]",
		},
	}

	for _, tst := range testingTable {
		r := references.CellReferenceMatch.ReplaceAllString(tst.from, fmt.Sprintf("%s$1", prefix))
		ExpectEquality(t, r, tst.to)
	}
}

func TestExpressionsAdjustment(t *testing.T) {
	type test struct {
		from string
		to   string
	}

	var testingTable []test = []test{
		{
			from: "{A1} + {B100}",
			to:   "{B2} + {C101}",
		},
		{
			from: "{A1} + {B100[1]} / {ZZZ2309[100][203]}",
			to:   "{B2} + {C101[1]} / {AAAA2310[100][203]}",
		},
	}

	adj := func(_ cells.Position) cells.Adjustment {
		return cells.Adjustment{
			Row:    1,
			Column: 1,
		}
	}

	for _, tst := range testingTable {
		s, err := references.AdjustReferencesInExpression(tst.from, adj)
		ExpectEquality(t, err, nil)
		ExpectEquality(t, s, tst.to)
	}
}

func TestExpressionsAdjustmentWithFilter(t *testing.T) {
	type test struct {
		from string
		to   string
	}

	var testingTable []test = []test{
		{
			from: "{A1} + {B100}",
			to:   "{A1} + {B100}",
		},
		{
			from: "{A1} + {B100[1]} / {ZZZ2310[100][203]}",
			to:   "{A1} + {B100[1]} / {ZZZ2321[100][203]}",
		},
	}

	adj := func(p cells.Position) cells.Adjustment {
		if p.Row > 1000 {
			return cells.Adjustment{
				Row:    11,
				Column: 0,
			}
		}
		return cells.Adjustment{}
	}

	for _, tst := range testingTable {
		s, err := references.AdjustReferencesInExpression(tst.from, adj)
		ExpectEquality(t, err, nil)
		ExpectEquality(t, s, tst.to)
	}
}
