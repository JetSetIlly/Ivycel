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

	statusBarHeight int
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
		worksheet.Size(-1, -1-float32(iv.statusBarHeight))

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

	var statusBar *giu.LabelWidget
	{
		lastErr := iv.ivy.LastError()
		if lastErr == nil {
			statusBar = giu.Label("Ready")
		} else {
			statusBar = giu.Label(lastErr.Error())
		}
	}

	inputBase, outputBase := iv.ivy.Base()

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
				giu.Menu("Base").Layout(
					giu.Label("Input"),
					giu.Spacing(),
					giu.MenuItem("binary").Selected(inputBase == 2).OnClick(func() {
						iv.ivy.SetBase(2, outputBase)
						iv.worksheet.RecalculateAll()
					}),
					giu.MenuItem("octal").Selected(inputBase == 8).OnClick(func() {
						iv.ivy.SetBase(8, outputBase)
						iv.worksheet.RecalculateAll()
					}),
					giu.MenuItem("decimal").Selected(inputBase == 10).OnClick(func() {
						iv.ivy.SetBase(10, outputBase)
						iv.worksheet.RecalculateAll()
					}),
					giu.MenuItem("hexadecimal").Selected(inputBase == 16).OnClick(func() {
						iv.ivy.SetBase(16, outputBase)
						iv.worksheet.RecalculateAll()
					}),
					giu.Spacing(),
					giu.Separator(),
					giu.Spacing(),
					giu.Label("Output"),
					giu.Spacing(),
					giu.MenuItem("binary").Selected(outputBase == 2).OnClick(func() {
						iv.ivy.SetBase(inputBase, 2)
						iv.worksheet.RecalculateAll()
					}),
					giu.MenuItem("octal").Selected(outputBase == 8).OnClick(func() {
						iv.ivy.SetBase(inputBase, 8)
						iv.worksheet.RecalculateAll()
					}),
					giu.MenuItem("decimal").Selected(outputBase == 10).OnClick(func() {
						iv.ivy.SetBase(inputBase, 10)
						iv.worksheet.RecalculateAll()
					}),
					giu.MenuItem("hexadecimal").Selected(outputBase == 16).OnClick(func() {
						iv.ivy.SetBase(inputBase, 16)
						iv.worksheet.RecalculateAll()
					}),
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

		// measure height of status bar
		giu.Custom(func() {
			iv.statusBarHeight = giu.GetCursorScreenPos().Y
		}),
		giu.Spacing(),
		statusBar,
		giu.Custom(func() {
			iv.statusBarHeight = giu.GetCursorScreenPos().Y - iv.statusBarHeight
		}),
	)
}

func main() {
	iv := ivycel{
		ivy: ivy.New(),
	}
	iv.worksheet = worksheet.NewWorksheet(&iv.ivy, 20, 20)
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
