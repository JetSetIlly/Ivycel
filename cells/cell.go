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

func (c *Cell) Commit() {
	c.Entry = strings.TrimSpace(c.Entry)
	if c.Entry == "" {
		c.result = ""
		c.err = nil
		return
	}
	r, err := c.engine.Execute(c.position.Reference(), c.Entry)
	if err != nil {
		c.err = err
		return
	}

	r = strings.TrimSpace(r)
	rowSplit := strings.Split(r, "\n")
	for ri, rv := range rowSplit {
		colSplit := strings.Fields(rv)

		var from int
		if ri == 0 {
			c.result = colSplit[0]
			from = 1
		}

		if from < len(colSplit) {
			for ci, cv := range colSplit[from:] {
				rel := c.worksheet.RelativeCell(c, Position{Row: ri, Column: ci + from})
				if rel == nil {
					break
				}
				rel.Entry = strings.TrimSpace(cv)
				rel.Commit()
			}
		}
	}
}

func (c *Cell) Result() string {
	return c.result
}
