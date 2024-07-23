package cells

import (
	"strings"

	"github.com/jetsetilly/ivycel/engine"
)

type Cell struct {
	engine   engine.Interface
	position Position

	Entry  string
	result string
	err    error

	User any
}

func NewCell(engine engine.Interface, position Position) *Cell {
	return &Cell{
		engine:   engine,
		position: position,
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
	c.result, c.err = c.engine.Execute(c.position.Reference(), c.Entry)
}

func (c *Cell) Result() string {
	return c.result
}
