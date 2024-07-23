package cells

import "strings"

type Engine interface {
	Execute(id string, ex string) (string, error)
}

type Cell struct {
	engine   Engine
	position Position

	Entry  string
	result string
	err    error

	User any
}

func NewCell(engine Engine, position Position) *Cell {
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