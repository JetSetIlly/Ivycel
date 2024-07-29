package worksheet

import (
	"fmt"
	"log"
	"math/rand"

	"github.com/jetsetilly/ivycel/cells"
	"github.com/jetsetilly/ivycel/engine"
	"github.com/jetsetilly/ivycel/references"
)

type User func(cell *cells.Cell)

type Worksheet struct {
	engine engine.Interface
	user   User

	rows    int
	columns int

	// current positions of cells references by cell ID
	positions       map[cells.CellID]cells.Position
	cellsByPosition map[cells.Position]cells.CellID
	cellsByID       map[cells.CellID]*cells.Cell

	User any
}

func NewWorksheet(engine engine.Interface, rows int, columns int, user User) Worksheet {
	ws := Worksheet{
		engine:          engine,
		user:            user,
		rows:            rows,
		columns:         columns,
		positions:       make(map[cells.CellID]cells.Position),
		cellsByPosition: make(map[cells.Position]cells.CellID),
		cellsByID:       make(map[cells.CellID]*cells.Cell),
	}

	for row := range ws.rows {
		for col := range ws.columns {
			ws.createCell(cells.Position{Row: row, Column: col})
		}
	}

	return ws
}

func (ws *Worksheet) createCell(pos cells.Position) {
	id := cells.CellID(fmt.Sprintf("cell%v", rand.Int63()))
	ws.positions[id] = pos
	ws.cellsByPosition[pos] = id
	cell := cells.NewCell(ws.engine, ws, id)
	ws.cellsByID[id] = cell
	ws.engine.Execute(pos.Reference(), "0")
	ws.user(cell)
}

func (ws Worksheet) Position(cell cells.CellID) cells.Position {
	return ws.positions[cell]
}

func (ws *Worksheet) adjustCells(adj func(p cells.Position) cells.Adjustment) {
	// rules common to any adjustment that need to be obeyed
	commonAdj := func(p cells.Position) cells.Adjustment {
		// use parent's position if the cell has one
		parent := ws.cellsByID[ws.cellsByPosition[p]].Parent()
		if parent != nil {
			p = parent.Position()
		}

		return adj(p)
	}

	// change expressions for all cells
	for rowi := range ws.rows {
		for coli := range ws.columns {
			pos := cells.Position{Row: rowi, Column: coli}
			id := ws.cellsByPosition[pos]
			cell := ws.cellsByID[id]

			var err error
			cell.Entry, err = references.AdjustReferencesInExpression(cell.Entry, commonAdj)
			if err != nil {
				log.Printf("worksheet: adjustCells: %s", err.Error())
			}

			cell.Commit(false)
		}
	}
}

func (ws *Worksheet) InsertRow(at int) {
	for rowi := ws.rows; rowi >= at; rowi-- {
		for coli := range ws.columns {
			pos := cells.Position{Row: rowi, Column: coli}
			id := ws.cellsByPosition[pos]
			pos.Row++
			ws.cellsByPosition[pos] = id
			ws.positions[id] = pos
		}
	}
	for coli := range ws.columns {
		ws.createCell(cells.Position{Row: at, Column: coli})
	}

	ws.rows++

	ws.adjustCells(func(p cells.Position) cells.Adjustment {
		if p.Row >= at {
			return cells.Adjustment{Row: 1}
		}
		return cells.Adjustment{}
	})
}

func (ws *Worksheet) InsertColumn(at int) {
	for coli := ws.columns; coli >= at; coli-- {
		for rowi := range ws.rows {
			pos := cells.Position{Row: rowi, Column: coli}
			id := ws.cellsByPosition[pos]
			pos.Column++
			ws.cellsByPosition[pos] = id
			ws.positions[id] = pos
		}
	}
	for rowi := range ws.rows {
		ws.createCell(cells.Position{Row: rowi, Column: at})
	}

	ws.columns++

	ws.adjustCells(func(p cells.Position) cells.Adjustment {
		if p.Column >= at {
			return cells.Adjustment{Column: 1}
		}
		return cells.Adjustment{}
	})
}

func (ws Worksheet) Cell(row int, column int) *cells.Cell {
	id := ws.cellsByPosition[cells.Position{Row: row, Column: column}]
	return ws.cellsByID[id]
}

func (ws Worksheet) Size() (int, int) {
	return ws.rows, ws.columns
}

func (ws Worksheet) RecalculateAll() {
	ws.engine.WithErrorSupression(func() {
		for rowi := range ws.rows {
			for coli := range ws.columns {
				id := ws.cellsByPosition[cells.Position{Row: rowi, Column: coli}]
				ws.cellsByID[id].Commit(false)
			}
		}
	})
}

func (ws Worksheet) RelativeCell(root *cells.Cell, pos cells.Position) *cells.Cell {
	pos.Row += root.Position().Row
	pos.Column += root.Position().Column
	if pos.Row >= ws.rows || pos.Column >= ws.columns {
		return nil
	}
	id := ws.cellsByPosition[pos]
	return ws.cellsByID[id]
}
