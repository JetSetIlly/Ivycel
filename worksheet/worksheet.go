package worksheet

import (
	"github.com/jetsetilly/ivycel/cells"
	"github.com/jetsetilly/ivycel/engine"
)

type Worksheet struct {
	engine engine.Interface
	cells  [][]*cells.Cell

	User any
}

func NewWorksheet(engine engine.Interface, rows int, columns int) Worksheet {
	ws := Worksheet{
		engine: engine,
	}

	ws.cells = make([][]*cells.Cell, rows)
	for i := 0; i < len(ws.cells); i++ {
		ws.cells[i] = make([]*cells.Cell, columns)
		for j := 0; j < len(ws.cells[i]); j++ {
			ws.cells[i][j] = cells.NewCell(ws.engine, ws, cells.Position{Row: i, Column: j})
			ws.engine.Execute(ws.cells[i][j].Position().Reference(), "0")
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
			ws.cells[i][j].Commit(false)
		}
	}
}

func (ws Worksheet) RelativeCell(root *cells.Cell, pos cells.Position) *cells.Cell {
	pos.Row += root.Position().Row
	pos.Column += root.Position().Column
	if pos.Row >= len(ws.cells) || pos.Column >= len(ws.cells[pos.Row]) {
		return nil
	}
	return ws.cells[pos.Row][pos.Column]
}
