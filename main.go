package main

import (
	"fmt"

	imgui "github.com/AllenDang/cimgui-go"
	"github.com/AllenDang/giu"
	"github.com/jetsetilly/ivycel/cells"
	"github.com/jetsetilly/ivycel/ivy"
	"github.com/jetsetilly/ivycel/worksheet"
)

type ivycel struct {
	ivy       ivy.Ivy
	worksheet worksheet.Worksheet
}

type worksheetUser struct {
	selected     *cells.Cell
	editing      *cells.Cell
	focusCell    bool
	focusFormula bool
}

type cellUser struct {
}

func (iv *ivycel) layout() {
	w := giu.SingleWindowWithMenuBar()

	var selected *giu.LabelWidget
	selected = giu.Label(iv.worksheet.User.(*worksheetUser).selected.Position().Reference())

	var formula *giu.InputTextWidget
	formula = giu.InputText(&iv.worksheet.User.(*worksheetUser).selected.Entry)
	formula.Flags(giu.InputTextFlagsEnterReturnsTrue)
	formula.OnChange(func() {
		iv.worksheet.User.(*worksheetUser).selected.Commit()
		iv.worksheet.RecalculateAll()
		iv.worksheet.User.(*worksheetUser).editing = nil
		iv.worksheet.User.(*worksheetUser).focusFormula = true
	})
	formula.Size(-1)

	var worksheet *giu.TableWidget
	{
		worksheet = giu.Table()

		rowCt, colCt := iv.worksheet.Size()

		var cols []*giu.TableColumnWidget
		c := giu.TableColumn("")
		cols = append(cols, c)
		for i := range colCt {
			c := giu.TableColumn(cells.NumericToBase26(i))
			c.Flags(giu.TableColumnFlagsWidthFixed)
			c.InnerWidthOrWeight(100)
			cols = append(cols, c)
		}

		var rowHeight float32
		{
			rowHeight = imgui.CalcTextSize("X").Y
			_, y := giu.GetItemInnerSpacing()
			rowHeight += y * 2
		}

		var rows []*giu.TableRowWidget
		for i := range rowCt {
			var rowCols []giu.Widget
			rowCols = append(rowCols, giu.Label(fmt.Sprintf("%d", i)))
			for j := range colCt {
				cell := iv.worksheet.CellEntry(i, j)

				if iv.worksheet.User.(*worksheetUser).editing == cell {
					giu.SetKeyboardFocusHere()
					inp := giu.InputText(&cell.Entry).Size(-1)
					inp.Flags(giu.InputTextFlagsEnterReturnsTrue | giu.InputTextFlagsAutoSelectAll)
					inp.OnChange(func() {
						iv.worksheet.User.(*worksheetUser).editing = nil
						cell.Commit()
						iv.worksheet.RecalculateAll()
					})
					rowCols = append(rowCols,
						giu.Row(
							inp,
							giu.Custom(func() {
								if iv.worksheet.User.(*worksheetUser).focusCell {
									iv.worksheet.User.(*worksheetUser).focusCell = false
									giu.SetKeyboardFocusHereV(-1)
								}
							}),
						),
					)
				} else {
					p := giu.GetCursorScreenPos()
					lab := giu.Label(cell.Result())
					ev := giu.Event().OnClick(giu.MouseButtonLeft, func() {
						iv.worksheet.User.(*worksheetUser).selected = cell
					}).OnDClick(giu.MouseButtonLeft, func() {
						iv.worksheet.User.(*worksheetUser).editing = cell
						iv.worksheet.User.(*worksheetUser).focusCell = true
					})
					giu.SetCursorScreenPos(p)

					inv := giu.InvisibleButton().Size(-1, rowHeight)
					ev2 := giu.Event().OnClick(giu.MouseButtonLeft, func() {
						iv.worksheet.User.(*worksheetUser).selected = cell
					}).OnDClick(giu.MouseButtonLeft, func() {
						if iv.worksheet.User.(*worksheetUser).editing != nil {
							iv.worksheet.User.(*worksheetUser).editing.Commit()
							iv.worksheet.RecalculateAll()
						}
						iv.worksheet.User.(*worksheetUser).editing = cell
						iv.worksheet.User.(*worksheetUser).focusCell = true
					})
					rowCols = append(rowCols, giu.Row(lab, ev, inv, ev2))
				}

			}
			rows = append(rows, giu.TableRow(rowCols...))
		}

		worksheet.Columns(cols...)
		worksheet.Rows(rows...)
		worksheet.Freeze(1, 1)
		worksheet.Flags(giu.TableFlagsBorders | giu.TableFlagsScrollY | giu.TableFlagsScrollX)
	}

	w.Layout(
		giu.Style().SetFontSize(16).To(
			giu.MenuBar().Layout(
				giu.Menu("File").Layout(
					giu.MenuItem("Open").Shortcut("Ctrl+O"),
					giu.MenuItem("Save"),
					giu.Menu("Save as ...").Layout(
						giu.MenuItem("Excel file"),
						giu.MenuItem("CSV file"),
					),
				),
				giu.Menu("Edit").Layout(
					giu.MenuItem("<placeholder>"),
				),
				giu.Menu("View").Layout(
					giu.MenuItem("<placeholder>"),
				),
				giu.Menu("Insert").Layout(
					giu.MenuItem("<placeholder>"),
				),
				giu.Menu("Format").Layout(
					giu.MenuItem("<placeholder>"),
				),
			),
		),
		giu.Style().SetFontSize(18).To(
			giu.Row(
				selected,
				giu.Custom(func() {
					if iv.worksheet.User.(*worksheetUser).focusFormula {
						giu.SetKeyboardFocusHere()
						iv.worksheet.User.(*worksheetUser).focusFormula = false
					}
				}),
				formula,
			),
			worksheet,
		),
	)
}

func main() {
	iv := ivycel{
		ivy: ivy.New(),
	}
	iv.worksheet = worksheet.NewWorksheet(iv.ivy, 20, 20)
	iv.worksheet.User = &worksheetUser{
		selected:     iv.worksheet.CellEntry(0, 0),
		focusFormula: true,
	}

	rowCt, colCt := iv.worksheet.Size()
	for i := range rowCt {
		for j := range colCt {
			iv.worksheet.CellEntry(i, j).User = &cellUser{}
		}
	}

	wnd := giu.NewMasterWindow("Ivycel", 800, 600, giu.MasterWindowFlagsNotResizable)
	wnd.Run(iv.layout)
}
