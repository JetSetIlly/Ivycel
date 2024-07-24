package main

import (
	"fmt"
	"image/color"

	imgui "github.com/AllenDang/cimgui-go"
	"github.com/AllenDang/giu"
	"github.com/jetsetilly/ivycel/cells"
	"github.com/jetsetilly/ivycel/engine/ivy"
	"github.com/jetsetilly/ivycel/worksheet"
)

type ivycel struct {
	ivy ivy.Ivy

	worksheet worksheet.Worksheet

	cellNormal   *giu.StyleSetter
	cellReadOnly *giu.StyleSetter
	cellEdit     *giu.StyleSetter

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

// the window menu is complicated enough to warrant its own function
func (iv *ivycel) layoutMenu() giu.Widget {
	inputBase, outputBase := iv.ivy.Base()

	inputBaseMenuItem := func(label string, base int) giu.Widget {
		return giu.MenuItem(label).Selected(inputBase == base).OnClick(func() {
			iv.ivy.SetBase(base, outputBase)
			iv.worksheet.RecalculateAll()
		})
	}

	outputBaseMenuItem := func(label string, base int) giu.Widget {
		return giu.MenuItem(label).Selected(outputBase == base).OnClick(func() {
			iv.ivy.SetBase(inputBase, base)
			iv.worksheet.RecalculateAll()
		})
	}

	return giu.Style().SetFontSize(16.5).To(
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
				inputBaseMenuItem("Binary", 2),
				inputBaseMenuItem("Octal", 8),
				inputBaseMenuItem("Decimal", 10),
				inputBaseMenuItem("Hexadecimal", 16),
				giu.Spacing(),
				giu.Separator(),
				giu.Spacing(),
				giu.Label("Output"),
				giu.Spacing(),
				outputBaseMenuItem("Binary", 2),
				outputBaseMenuItem("Octal", 8),
				outputBaseMenuItem("Decimal", 10),
				outputBaseMenuItem("Hexadecimal", 16),
			),
		),
	)
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

	// the main body of the spreadsheet is a table
	var worksheet *giu.TableWidget
	{
		var cols []*giu.TableColumnWidget
		var rows []*giu.TableRowWidget

		rowCt, colCt := iv.worksheet.Size()

		worksheet = giu.Table()
		worksheet.Size(-1, -1-float32(iv.statusBarHeight))
		worksheet.Freeze(1, 1)
		worksheet.Flags(giu.TableFlagsScrollY | giu.TableFlagsScrollX | giu.TableFlagsResizable)

		// first column is the row number and should not be resizeable
		cols = append(cols, giu.TableColumn("").Flags(giu.TableColumnFlagsNoResize))

		// remaining columns are worksheet columns numbered from "A"
		for i := range colCt {
			c := giu.TableColumn(cells.NumericToBase26(i))
			c.Flags(giu.TableColumnFlagsWidthFixed)
			c.InnerWidthOrWeight(100)
			cols = append(cols, c)
		}

		// height of each row
		rowHeight := imgui.CalcTextSize("X").Y
		_, y := giu.GetItemInnerSpacing()
		rowHeight += y * 2

		for i := range rowCt {
			var rowCols []giu.Widget

			// first column of each row is the row number
			rowCols = append(rowCols, giu.Custom(func() {
				giu.AlignTextToFramePadding()
				giu.Label(fmt.Sprintf(" %d", i+1)).Build()
			}))

			for j := range colCt {
				// reference to the cell at row/column number
				cell := iv.worksheet.CellEntry(i, j)

				// how we display the cell depends on whether the cell is the
				// one currently being edited
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
						giu.Custom(func() {
							iv.cellEdit.Push()
							defer iv.cellEdit.Pop()
							inp.Build()
							if iv.worksheet.User.(*worksheetUser).focusCell {
								iv.worksheet.User.(*worksheetUser).focusCell = false
								giu.SetKeyboardFocusHereV(-1)
							}
						}),
					)
				} else {
					// each cell is a button with a tooltip. the tooltip can be
					// an empty widget meaning that it will never appear
					var cel *giu.ButtonWidget
					var tip giu.Widget

					if err := cell.Error(); err != nil {
						cel = giu.Button("???")
						tip = giu.Tooltip(err.Error())
					} else {
						cel = giu.Button(cell.Result())
						tip = giu.Custom(func() {})
					}

					// each cell is the width of the column it is in and the
					// height of the row
					cel.Size(-1, rowHeight)

					// event handler for cell deals with mouse clicks. we prefer
					// this to the Button.OnClick() functio
					var ev *giu.EventHandler
					ev = giu.Event()

					ev.OnClick(giu.MouseButtonLeft, func() {
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
					})

					ev.OnDClick(giu.MouseButtonLeft, func() {
						if iv.worksheet.User.(*worksheetUser).editing != nil {
							iv.worksheet.User.(*worksheetUser).editing.Commit(true)
							iv.worksheet.RecalculateAll()
						}
						if !cell.ReadOnly() {
							iv.worksheet.User.(*worksheetUser).editing = cell
							iv.worksheet.User.(*worksheetUser).focusCell = true
						}
					})

					// decide on display style for cell
					var sty *giu.StyleSetter
					if cell.ReadOnly() {
						sty = iv.cellReadOnly
					} else {
						sty = iv.cellNormal

					}

					rowCols = append(rowCols,
						giu.Custom(func() {
							sty.Push()
							defer sty.Pop()
							cel.Build()
							ev.Build()
							tip.Build()
						}))
				}

			}
			rows = append(rows, giu.TableRow(rowCols...))
		}

		worksheet.Columns(cols...)
		worksheet.Rows(rows...)
	}

	// the status bar appears at the bottom of the window. the height of the
	// status bar is measured during the layout directive below
	var statusBar *giu.LabelWidget
	{
		lastErr := iv.ivy.LastError()
		if lastErr == nil {
			statusBar = giu.Label("Ready")
		} else {
			statusBar = giu.Label(lastErr.Error())
		}
	}

	w.Layout(
		iv.layoutMenu(),
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

func (iv *ivycel) setStyling() {
	iv.cellNormal = giu.Style().
		SetStyleFloat(giu.StyleVarFrameBorderSize, 0).
		SetStyleFloat(giu.StyleVarFrameRounding, 0).
		SetStyle(giu.StyleVarButtonTextAlign, 0, 0).
		SetColor(giu.StyleColorText, color.RGBA{R: 255, G: 255, B: 255, A: 255})

	iv.cellReadOnly = giu.Style().
		SetStyleFloat(giu.StyleVarFrameBorderSize, 0).
		SetStyleFloat(giu.StyleVarFrameRounding, 0).
		SetStyle(giu.StyleVarButtonTextAlign, 0, 0).
		SetColor(giu.StyleColorButton, color.Transparent)

	iv.cellEdit = giu.Style().
		SetStyleFloat(giu.StyleVarFrameBorderSize, 2).
		SetStyleFloat(giu.StyleVarFrameRounding, 3).
		SetColor(giu.StyleColorBorder, color.RGBA{R: 100, G: 100, B: 200, A: 255})
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
	iv.setStyling()
	wnd.Run(iv.layout)
}
