package ivy

import (
	"fmt"
	"regexp"
)

var CellReferenceMatch = regexp.MustCompile("{([[:alpha:]]+[[:digit:]]+)}")

var normalisedCellReferencePrefix = "v"

// normalise cell references so they can be used inside of ivy
func normaliseCellReferences(ref string, ex string) (string, string) {
	ref = fmt.Sprintf("%s%s", normalisedCellReferencePrefix, ref)
	ex = CellReferenceMatch.ReplaceAllString(ex, fmt.Sprintf("%s$1", normalisedCellReferencePrefix))
	return ref, ex
}

// wrap cell reference so that it is safe to use with ivy in all instances. it only needs to be
// called when the cell appears in an ivy expression
func WrapCellReference(ref string) string {
	return fmt.Sprintf("{%s}", ref)
}
