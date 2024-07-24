package cells

import (
	"strings"

	"github.com/jetsetilly/ivycel/engine"
)

type Worksheet interface {
	RelativeCell(root *Cell, pos Position) *Cell
}

type Cell struct {
	engine    engine.Interface
	worksheet Worksheet
	position  Position

	Entry  string
	result string
	err    error

	parent   *Cell
	children []*Cell

	User any
}

func NewCell(engine engine.Interface, worksheet Worksheet, position Position) *Cell {
	return &Cell{
		engine:    engine,
		worksheet: worksheet,
		position:  position,
	}
}

func (c Cell) Position() Position {
	return c.position
}

// force argument will commit the cell even if the cell is read-only. this is
// normally what you would want unless you were re-committing as part of a
// recalculation
func (c *Cell) Commit(force bool) {
	if c.ReadOnly() && !force {
		return
	}

	// clear previous results from child cells
	for _, child := range c.children {
		child.Entry = ""
		child.Commit(true)
	}

	// if entry is empty then we don't need to do any more except tidy up
	c.Entry = strings.TrimSpace(c.Entry)
	if c.Entry == "" {
		_, _ = c.engine.Execute(c.position.Reference(), "0")
		c.result = ""
		c.err = nil
		c.parent = nil
		return
	}

	// execute contents of cell
	r, err := c.engine.Execute(c.position.Reference(), c.Entry)
	if err != nil {
		c.err = err
		return
	}

	inputBase, outputBase := c.engine.Base()
	c.engine.SetBase(outputBase, outputBase)
	defer c.engine.SetBase(inputBase, outputBase)

	r = strings.TrimSpace(r)
	rowSplit := strings.Split(r, "\n")
	for ri, rv := range rowSplit {
		colSplit := strings.Fields(rv)

		var from int
		if ri == 0 {
			c.result = colSplit[0]
			c.parent = nil
			from = 1
		}

		if from < len(colSplit) {
			for ci, cv := range colSplit[from:] {
				rel := c.worksheet.RelativeCell(c, Position{Row: ri, Column: ci + from})
				if rel == nil {
					break
				}

				c.children = append(c.children, rel)

				rel.Entry = strings.TrimSpace(cv)
				rel.parent = c
				rel.result, rel.err = c.engine.Execute(rel.position.Reference(), rel.Entry)
			}
		}
	}
}

func (c *Cell) Result() string {
	return c.result
}

func (c *Cell) ReadOnly() bool {
	return c.parent != nil
}
