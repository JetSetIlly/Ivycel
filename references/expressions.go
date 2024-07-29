package references

import (
	"strings"

	"github.com/jetsetilly/ivycel/cells"
)

// adjust all cell references in expression by an adjustment, which is provided
// by the adj function. the adj function takes the position of each reference
// in the expression and returns the amount that the cell should be adjusted by
//
// for example: the expression has a reference to cell A100. the provided adj
// function tests that this reference is to a cell somewhere in row 50 or above.
// because A100 satisfies that test a cells.Adjustment value of {Row: 1, Column:
// 0} is returned. this will cause the cell reference to be adjusted to A101
//
// cell references in the expression must be wrapped for them to be considered
// for adjustment
//
// in case of error the unadjusted expression is returned
func AdjustReferencesInExpression(expression string, adj func(cells.Position) cells.Adjustment) (string, error) {

	mtchs := CellReferenceMatch.FindAllStringSubmatch(expression, -1)

	type mapping struct {
		from string
		to   string
	}

	var replacements []mapping

	for _, m := range mtchs {
		ref := m[referenceWithoutIndex]
		p, err := cells.PositionFromReference(ref)
		if err != nil {
			return expression, err
		}

		adjRef, err := AdjustReference(ref, adj(p))
		if err != nil {
			return expression, err
		}

		replacements = append(replacements, mapping{
			from: ref,
			to:   strings.Replace(m[referenceWithoutIndex], ref, adjRef, 1),
		})
	}

	for _, r := range replacements {
		expression = strings.Replace(expression, r.from, r.to, -1)
	}

	return expression, nil
}
