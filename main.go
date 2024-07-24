package main

import (
	"fmt"

	imgui "github.com/AllenDang/cimgui-go"
	"github.com/AllenDang/giu"
	"github.com/jetsetilly/ivycel/cells"
	"github.com/jetsetilly/ivycel/engine/ivy"
	"github.com/jetsetilly/ivycel/worksheet"
)

type ivycel struct {
	ivy       ivy.Ivy
	worksheet worksheet.Worksheet

	statusBarHeight int
}

type worksheetUser struct {
	selected *cells.Cell
	editing  *cells.Cell

	// focus either the cell-input or the formula-input on the next update
	focusCell    bool
	focusFormula bool
}

type cellUser struct {
	// the edit cursor position is stored for each cell. however, because only
	// one cell can be in the 'editing' mode at once, we can probably keep this
	// value in the ivycel type. but this seems more correct
	editCursorPosition int
}

func (iv *ivycel) layout() {
	w := giu.SingleWindowWithMenuBar()

	var selected *giu.LabelWidget
	selected = giu.Label(iv.worksheet.User.(*worksheetUser).selected.Position().Reference())

	var formula *giu.InputTextWidget
	{
		formula = giu.InputText(&iv.worksheet.User.(*worksheetUser).selected.Entry)
		formula.Flags(giu.InputTextFlagsEnterReturnsTrue)
		formula.OnChange(func() {
			iv.worksheet.User.(*worksheetUser).selected.Commit(true)
			iv.worksheet.RecalculateAll()
			iv.worksheet.User.(*worksheetUser).editing = nil
			iv.worksheet.User.(*worksheetUser).focusFormula = true
		})
		formula.Size(-1)
	}

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
			rowCols = append(rowCols, giu.Label(fmt.Sprintf("%d", i+1)))
			for j := range colCt {
				cell := iv.worksheet.CellEntry(i, j)

				if iv.worksheet.User.(*worksheetUser).editing == cell {
					giu.SetKeyboardFocusHere()
					inp := giu.InputText(&cell.Entry).Size(-1)

					// CalbackAlways flag so we can update the editCursorPosition every keypress
					// and EnterReturnsTrue so that OnChange() is not triggered until editing
					// has finished
					inp.Flags(giu.InputTextFlagsCallbackAlways | giu.InputTextFlagsEnterReturnsTrue)

					// keep track of current cursor position in the input
					// widget. we use this to insert cell references at the
					// correct point
					inp.Callback(func(data imgui.InputTextCallbackData) int {
						cell.User.(*cellUser).editCursorPosition = int(data.CursorPos())
						return 0
					})

					inp.OnChange(func() {
						iv.worksheet.User.(*worksheetUser).editing = nil
						cell.Commit(true)
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
					// label and invisible button necessary to make as much of
					// the cell area as clickable as possible

					onClick := func() {
						if iv.worksheet.User.(*worksheetUser).editing != nil {
							// insert selected cell reference to cell being edited
							editCell := iv.worksheet.User.(*worksheetUser).editing
							pos := editCell.User.(*cellUser).editCursorPosition
							editCell.Entry = fmt.Sprintf("%s%s%s",
								editCell.Entry[:pos],
								iv.ivy.WrapCellReference(cell.Position().Reference()),
								editCell.Entry[pos:],
							)

							// cell being editing will need to be refocused after the user click
							iv.worksheet.User.(*worksheetUser).focusCell = true
						} else {
							iv.worksheet.User.(*worksheetUser).selected = cell
						}
					}

					onDClick := func() {
						if iv.worksheet.User.(*worksheetUser).editing != nil {
							iv.worksheet.User.(*worksheetUser).editing.Commit(true)
							iv.worksheet.RecalculateAll()
						}
						if !cell.ReadOnly() {
							iv.worksheet.User.(*worksheetUser).editing = cell
							iv.worksheet.User.(*worksheetUser).focusCell = true
						}
					}

					var tooltip giu.Widget
					tooltip = giu.Custom(func() {})

					p := giu.GetCursorScreenPos()
					var lab *giu.LabelWidget
					if err := cell.Error(); err != nil {
						lab = giu.Label("???")
						tooltip = giu.Tooltip(err.Error())
					} else {
						lab = giu.Label(cell.Result())
					}
					evLab := giu.Event().OnClick(giu.MouseButtonLeft, onClick)
					evLab.OnDClick(giu.MouseButtonLeft, onDClick)

					giu.SetCursorScreenPos(p)

					inv := giu.InvisibleButton().Size(-1, rowHeight)
					evInv := giu.Event().OnClick(giu.MouseButtonLeft, onClick)
					evInv.OnDClick(giu.MouseButtonLeft, onDClick)
					rowCols = append(rowCols, giu.Row(lab, evLab, tooltip, inv, evInv, tooltip))
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
				giu.Custom(func() {
					if iv.worksheet.User.(*worksheetUser).selected.ReadOnly() {
						giu.Style().SetDisabled(true).Push()
						defer giu.Style().SetDisabled(true).Pop()
					}
					formula.Build()
				}),
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
