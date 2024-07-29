package references

import (
	"fmt"
	"regexp"

	"github.com/jetsetilly/ivycel/cells"
)

// match anything inside paired braces that begins with a sequence of letters
// and then a sequence of digits. spaces not allowed at all
var CellReferenceMatch = regexp.MustCompile("{(([[:alpha:]]+[[:digit:]]+)(?U:[[:^space:]]*))}")

// note about the regex: the non-capturing group around the "[[:^space:]]*" form
// has the ungreedy flag set. this is because a greedy match would causes
// problem with an expression like "{A1}+{A2}". there are no spaces between the
// "{A1" and the closing brace after "A2" and so the match with "[[:^space:]]*
// would be on "}+{A2"

// list of match positions for CallReferenceMatch
const (
	directMatch           = 0
	unwrappedReference    = 1
	referenceWithoutIndex = 2
)

var normalisedCellReferencePrefix = "v"

// normalise cell references so they can be used inside of ivy
func NormaliseCellReferences(ref string, ex string) (string, string) {
	ref = fmt.Sprintf("%s%s", normalisedCellReferencePrefix, ref)
	ex = CellReferenceMatch.ReplaceAllString(ex, fmt.Sprintf("%s$1", normalisedCellReferencePrefix))
	return ref, ex
}

// wrap cell reference so that it is safe to use with ivy in all instances. it only needs to be
// called when the cell appears in an ivy expression
func WrapCellReference(ref string) string {
	return fmt.Sprintf("{%s}", ref)
}

// adjust cell reference by provided adjustment. cell references should not be
// wrapped. in case of error the unadjusted reference is returned
func AdjustReference(ref string, adj cells.Adjustment) (string, error) {
	p, err := cells.PositionFromReference(ref)
	if err != nil {
		return ref, err
	}
	p = p.Adjust(adj)
	return p.Reference(), nil
}
