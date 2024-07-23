package worksheet

import (
	"github.com/jetsetilly/ivycel/cells"
)

type Engine interface {
	Execute(id string, ex string) (string, error)
}

type Worksheet struct {
	engine Engine
	cells  [][]*cells.Cell

	User any
}

func NewWorksheet(engine Engine, rows int, columns int) Worksheet {
	ws := Worksheet{
		engine: engine,
	}

	ws.cells = make([][]*cells.Cell, rows)
	for i := 0; i < len(ws.cells); i++ {
		ws.cells[i] = make([]*cells.Cell, columns)
		for j := 0; j < len(ws.cells[i]); j++ {
			ws.cells[i][j] = cells.NewCell(ws.engine, cells.Position{Row: i, Column: j})
		}
	}

	return ws
}

func (ws Worksheet) CellEntry(row int, column int) *cells.Cell {
	return ws.cells[row][column]
}

func (ws Worksheet) Size() (int, int) {
	return len(ws.cells), len(ws.cells[0])
}

func (ws Worksheet) RecalculateAll() {
	for i := 0; i < len(ws.cells); i++ {
		for j := 0; j < len(ws.cells[i]); j++ {
			ws.cells[i][j].Commit()
		}
	}
}