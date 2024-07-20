package cells

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type Position struct {
	Row    int
	Column int
}

func (p Position) IsError() bool {
	return p.Row < 0 || p.Column < 0
}

func (p Position) String() string {
	return p.Reference()
}

// NumericToBase26 converts an integer into a base-26 number. Useful for
// generating column numbers for a spreadsheet.
func NumericToBase26(val int) string {
	const alphabetSize = 26

	var s string
	for val >= 0 {
		s = fmt.Sprintf("%c%s", ('A' + rune(val%alphabetSize)), s)
		val = (val / alphabetSize) - 1
	}
	return s
}

func (p Position) Reference() string {
	if p.IsError() {
		return ""
	}
	return fmt.Sprintf("%s%d", NumericToBase26(p.Column), p.Row)
}

var IllegalReference = errors.New("illegal position reference")

var errorPosition = Position{
	Row: -1, Column: -1,
}

func PositionFromReference(ref string) (Position, error) {
	ref = strings.TrimSpace(strings.ToUpper(ref))

	var col int
	var i int
	col = -1
	for i = 0; i < len(ref); i++ {
		c := rune(ref[i])
		if !(c >= 'A' && c <= 'Z') {
			break // for loop
		}
		c = c - rune('A')
		col = ((col + 1) * 26) + int(c)
	}

	if i == 0 {
		return errorPosition, fmt.Errorf("%w: no row number", IllegalReference)
	}
	ref = ref[i:]

	row, err := strconv.Atoi(ref)
	if err != nil {
		return errorPosition, fmt.Errorf("%w: malformed row number", IllegalReference)
	}

	return Position{Column: col, Row: row}, nil
}
