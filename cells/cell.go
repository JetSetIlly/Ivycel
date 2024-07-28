package cells

import (
	"errors"
	"strings"
	"unicode/utf8"

	"github.com/jetsetilly/ivycel/engine"
)

var UnsupportedShape = errors.New("unsupported shape")

type CellID string

type Worksheet interface {
	RelativeCell(root *Cell, pos Position) *Cell
	Position(CellID) Position
}

type Cell struct {
	engine    engine.Interface
	worksheet Worksheet
	id        CellID

	base engine.Base

	Entry  string
	result string
	err    error

	parent   *Cell
	children []*Cell

	User any
}

func NewCell(engine engine.Interface, worksheet Worksheet, id CellID) *Cell {
	return &Cell{
		id:        id,
		engine:    engine,
		worksheet: worksheet,
		base:      engine.Base(),
	}
}

func (c Cell) ID() CellID {
	return c.id
}

func (c Cell) Position() Position {
	return c.worksheet.Position(c.id)
}

// force argument will commit the cell even if the cell is read-only. this is
// normally what you would want unless you were re-committing as part of a
// recalculation. it also causes the engine to execute a zero assignment for the
// cell in the event of the Entry field being empty
func (c *Cell) Commit(force bool) {
	if c.ReadOnly() && !force {
		return
	}

	// clear previous results from child cells
	for _, child := range c.children {
		child.Entry = ""
		child.Commit(true)
	}
	c.children = c.children[:0]

	// reset other fields
	c.result = ""
	c.err = nil
	c.parent = nil

	// if entry is empty then we don't need to do any more except tidy up
	c.Entry = strings.TrimSpace(c.Entry)
	if c.Entry == "" {
		if force {
			_, _ = c.engine.Execute(c.Position().Reference(), "0")
		}
		return
	}

	var err error
	var r string

	// execute contents of cell
	c.engine.WithNumberBase(c.base, func() {
		r, err = c.engine.Execute(c.Position().Reference(), c.Entry)
	})
	if err != nil {
		c.err = err
		return
	}

	r = strings.TrimSpace(r)

	// do nothing if there are results
	colSplit := strings.Fields(r)
	if len(colSplit) == 0 {
		return
	}

	rowSplit := strings.Split(r, "\n")

	// check that we can work with the output. for example, an expression
	// like "2 2 2 2 rho iota 10" produces output that is too complicated
	for _, rsplt := range rowSplit {
		d, _ := utf8.DecodeRuneInString(rsplt)
		if d == '[' {
			c.err = UnsupportedShape
			return
		}
	}

	// when fill in relative cells we sometimes need to know the number of
	// columns there will ideally be in the output
	//
	// for example: the ivy expression "2 2 2 rho iota 10" will result in two
	// 2x2 matrices separated by a blank line. but in the spreadsheet output we
	// want the blank line to be represented by two empty & read-only cells. the
	// count of two is taken from the previous line's output
	var mostRecentColumnCount int

	group := 0
	for ri, rv := range rowSplit {
		colSplit := strings.Fields(rv)

		// handle blank lines
		if len(colSplit) == 0 {
			group++
			for ci := range mostRecentColumnCount {
				rel := c.worksheet.RelativeCell(c, Position{Row: ri, Column: ci})
				if rel == nil {
					break // for loop
				}

				c.children = append(c.children, rel)
				rel.Entry = ""
				rel.parent = c
				rel.result = ""
				rel.err = nil
			}
			continue // for rowSplit loop
		}

		// result line has columns

		// use this row's column count for next blank line
		mostRecentColumnCount = len(colSplit)

		// treat first column of first row differently
		var adj int
		if ri == 0 {
			c.result = colSplit[0]
			c.parent = nil
			adj = 1
		}

		// work through the columns in the row, taking into account the
		// adjustument (adj) required by the first row
		for ci, cv := range colSplit[adj:] {
			rel := c.worksheet.RelativeCell(c, Position{Row: ri, Column: ci + adj})
			if rel == nil {
				break // for loop
			}

			c.children = append(c.children, rel)
			rel.Entry = strings.TrimSpace(cv)
			rel.parent = c
			c.engine.WithErrorSupression(func() {
				c.engine.WithNumberBase(c.base.OutputOnly(), func() {
					rel.result, rel.err = c.engine.Execute(rel.Position().Reference(), rel.Entry)
				})
			})
		}
	}
}

func (c *Cell) Parent() *Cell {
	return c.parent
}

func (c *Cell) HasChildren() bool {
	return len(c.children) > 0
}

func (c *Cell) Result() string {
	return c.result
}

func (c *Cell) Error() error {
	return c.err
}

// if cell has a parent then it should be treated as read-only
func (c *Cell) ReadOnly() bool {
	return c.parent != nil
}

func (c *Cell) Base() engine.Base {
	if c.parent != nil {
		return c.parent.base
	}
	return c.base
}

func (c *Cell) SetBase(b engine.Base) {
	if c.parent != nil {
		c.parent.base = b
		return
	}
	c.base = b
}

// return the index for the root value of the cell depending on the shape of the value
func (c *Cell) RootIndex() string {
	if !c.HasChildren() {
		return ""
	}

	shape := c.engine.Shape(c.Position().Reference())
	n := len(strings.Fields(shape))

	// the Repeat() function assumes that we're counting from 1 inside the
	// engine. this is the default but maybe we should account for that possibly
	// being changed
	return strings.Repeat("[1]", n)
}
