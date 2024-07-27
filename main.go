package main

import (
	"fmt"
	"image"
	"image/color"

	imgui "github.com/AllenDang/cimgui-go"
	"github.com/AllenDang/giu"
	"github.com/jetsetilly/ivycel/cells"
	"github.com/jetsetilly/ivycel/engine"
	"github.com/jetsetilly/ivycel/engine/ivy"
	"github.com/jetsetilly/ivycel/fonts"
	"github.com/jetsetilly/ivycel/worksheet"
)

type ivycel struct {
	ivy ivy.Ivy

	worksheet worksheet.Worksheet

	cellNormal      *giu.StyleSetter
	cellReadOnly    *giu.StyleSetter
	cellEdit        *giu.StyleSetter
	outputBaseBadge *giu.StyleSetter

	statusBarHeight int

	boldFont *giu.FontInfo

	// fonts should be prepared as soon as possible in order for them to
	// be ready on the frame they are required. without preloading the
	// loading may be visible to the user
	fontsPreloaded bool
}

type worksheetUser struct {
	selected *cells.Cell
	editing  *cells.Cell

	// focus either the cell being edited or the formula bar on the next update
	focusCell    bool
	focusFormula bool
}

type cellUser struct {
	// the edit cursor position is stored for each cell. however, because only
	// one cell can be in the 'editing' mode at once, we can probably keep this
	// value in the ivycel type. but this seems more correct
	editCursorPosition int
}

// make sure the required fonts have been loaded at the correct size
func (iv *ivycel) preloadFonts() {
	if iv.fontsPreloaded {
		return
	}
	iv.fontsPreloaded = true

	giu.Style().
		SetFont(iv.boldFont).
		SetFontSize(fonts.BadgeFontSize).
		SetColor(giu.StyleColorText, color.Transparent).
		To(giu.Label(""))
}

// the window menu is complicated enough to warrant its own function
func (iv *ivycel) layoutMenu() giu.Widget {
	bs := iv.worksheet.User.(*worksheetUser).selected.Base()

	inputBaseMenuItem := func(label string, nbs int) giu.Widget {
		return giu.MenuItem(label).Selected(bs.Input == nbs).OnClick(func() {
			iv.worksheet.User.(*worksheetUser).selected.SetBase(engine.Base{Input: nbs, Output: bs.Output})
			iv.worksheet.RecalculateAll()
		})
	}

	outputBaseMenuItem := func(label string, nbs int) giu.Widget {
		return giu.MenuItem(label).Selected(bs.Output == nbs).OnClick(func() {
			iv.worksheet.User.(*worksheetUser).selected.SetBase(engine.Base{Input: bs.Input, Output: nbs})
			iv.worksheet.RecalculateAll()
		})
	}

	return giu.MenuBar().Layout(
		giu.Spacing(),
		giu.Menu(string(fonts.FileMenu)).Layout(
			giu.Label("File"),
			giu.Separator(),
			giu.MenuItem("Open..."),
			giu.MenuItem("Save..."),
		),
		giu.Spacing(),
		giu.Menu(string(fonts.InputBase)).Layout(
			giu.Label("Input Base"),
			giu.Separator(),
			inputBaseMenuItem("Binary", 2),
			inputBaseMenuItem("Octal", 8),
			inputBaseMenuItem("Decimal", 10),
			inputBaseMenuItem("Hexadecimal", 16),
		),
		giu.Menu(string(fonts.OutputBase)).Layout(
			giu.Label("Output Base"),
			giu.Separator(),
			outputBaseMenuItem("Binary", 2),
			outputBaseMenuItem("Octal", 8),
			outputBaseMenuItem("Decimal", 10),
			outputBaseMenuItem("Hexadecimal", 16),
		),
	)
}

func (iv *ivycel) layout() {
	iv.preloadFonts()

	var selected *giu.LabelWidget
	selected = giu.Label(iv.worksheet.User.(*worksheetUser).selected.Position().Reference())

	var formula *giu.InputTextWidget
	formula = giu.InputText(&iv.worksheet.User.(*worksheetUser).selected.Entry).
		Flags(giu.InputTextFlagsEnterReturnsTrue).
		OnChange(func() {
			iv.worksheet.User.(*worksheetUser).selected.Commit(true)
			iv.worksheet.RecalculateAll()
			iv.worksheet.User.(*worksheetUser).editing = nil
			iv.worksheet.User.(*worksheetUser).focusFormula = true
		}).
		Size(-1)

	// the main body of the spreadsheet is a table
	var worksheet *giu.TableWidget
	{
		rowCt, colCt := iv.worksheet.Size()

		worksheet = giu.Table().
			Size(-1, -1-float32(iv.statusBarHeight)).
			Freeze(1, 1).
			Flags(giu.TableFlagsScrollY | giu.TableFlagsScrollX | giu.TableFlagsResizable).
			NoHeader(true)

		// prepare columns for adding to table
		var cols []*giu.TableColumnWidget

		// first column is the row number and should not be resizeable
		cols = append(cols, giu.TableColumn("").Flags(giu.TableColumnFlagsNoResize))

		// remaining columns are worksheet columns numbered from "A"
		for range colCt {
			c := giu.TableColumn("").
				Flags(giu.TableColumnFlagsWidthFixed).
				InnerWidthOrWeight(100)
			cols = append(cols, c)
		}

		// add columns to table
		worksheet.Columns(cols...)

		// height of each row if fixed
		rowHeight := imgui.CalcTextSize("X").Y
		_, y := giu.GetItemInnerSpacing()
		rowHeight += y * 2

		// prepare rows for adding to table
		var rows []*giu.TableRowWidget

		{ // add column headers manually
			var rowCols []giu.Widget
			rowCols = append(rowCols, giu.Label(""))
			for i := range colCt {
				rowCols = append(rowCols, giu.Label(cells.NumericToBase26(i)))
			}
			rows = append(rows, giu.TableRow(rowCols...))
		}

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

				var badge giu.Widget
				if !cell.ReadOnly() {
					bs := cell.Base()
					if bs != iv.ivy.Base() {
						badge = giu.Custom(func() {
							iv.outputBaseBadge.Push()
							defer iv.outputBaseBadge.Pop()
							txt := fmt.Sprintf("%d", bs.Output)
							giu.SameLine()
							pos := giu.GetCursorScreenPos()
							pos = pos.Sub(image.Point{X: int(imgui.CalcTextSize(txt).X) + 10, Y: 2})
							giu.SetCursorScreenPos(pos)
							giu.Button(txt).Build()
						})
					}
				}

				// how we display the cell depends on whether the cell is the
				// one currently being edited
				if iv.worksheet.User.(*worksheetUser).editing == cell {
					giu.SetKeyboardFocusHere()
					inp := giu.InputText(&cell.Entry).Size(-1)

					// escape key cancels changes and deactivates the input text
					// for the cell
					if giu.IsKeyPressed(giu.KeyEscape) {
						iv.worksheet.User.(*worksheetUser).editing = nil
					}

					// CalbackAlways flag so we can update the editCursorPosition every keypress
					// and EnterReturnsTrue so that OnChange() is not triggered until editing
					// has finished
					inp.Flags(giu.InputTextFlagsCallbackAlways | giu.InputTextFlagsEnterReturnsTrue)

					// keep track of current cursor position in the input
					// widget. we use this to insert cell references at the
					// correct point
					inp.Callback(func(data imgui.InputTextCallbackData) int {
						if iv.worksheet.User.(*worksheetUser).focusCell {
							iv.worksheet.User.(*worksheetUser).focusCell = false
							data.SetCursorPos(int32(cell.User.(*cellUser).editCursorPosition))
							data.ClearSelection()
						}
						cell.User.(*cellUser).editCursorPosition = int(data.CursorPos())
						return 0
					})

					// on change function is only called on "enter returns true"
					// commit changes
					inp.OnChange(func() {
						iv.worksheet.User.(*worksheetUser).editing = nil
						cell.Commit(true)
						iv.worksheet.RecalculateAll()
					})

					rowCols = append(rowCols,
						giu.Custom(func() {
							iv.cellEdit.Push()
							defer iv.cellEdit.Pop()
							if iv.worksheet.User.(*worksheetUser).focusCell {
								// focusCell flag will be reset in the input widget's callback
								// function above. we do this because setting the keyboard focus
								// selects the entire contents of the input and we don't want that.
								// delaying the flag reset allows us to clear the selection and move
								// the input cursor
								giu.SetKeyboardFocusHere()
							}
							inp.Build()
							if badge != nil {
								badge.Build()
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
						if iv.worksheet.User.(*worksheetUser).editing == nil {
							iv.worksheet.User.(*worksheetUser).selected = cell
						}
					})

					ev.OnDClick(giu.MouseButtonLeft, func() {
						if iv.worksheet.User.(*worksheetUser).editing != nil {
							// insert selected cell reference to cell being edited
							editCell := iv.worksheet.User.(*worksheetUser).editing
							pos := editCell.User.(*cellUser).editCursorPosition
							ref := ivy.WrapCellReference(cell.Position().Reference())

							editCell.Entry = fmt.Sprintf("%s%s%s",
								editCell.Entry[:pos],
								ref,
								editCell.Entry[pos:],
							)

							// advance cursor position of edit cell
							editCell.User.(*cellUser).editCursorPosition += len(ref)

							// cell being editing will need to be refocused after double-click
							iv.worksheet.User.(*worksheetUser).focusCell = true
						} else if !cell.ReadOnly() {
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
							if badge != nil {
								badge.Build()
							}
						}))
				}

			}
			rows = append(rows, giu.TableRow(rowCols...))
		}

		// add rows to table
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

	giu.SingleWindowWithMenuBar().Layout(
		iv.layoutMenu(),
		giu.Style().SetFontSize(fonts.WorksheetFontSize).To(
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

	col := color.RGBA{R: 255, G: 100, B: 100, A: 200}
	iv.outputBaseBadge = giu.Style().
		SetFont(iv.boldFont).
		SetFontSize(fonts.BadgeFontSize).
		SetStyle(giu.StyleVarFramePadding, 2, 2).
		SetStyleFloat(giu.StyleVarFrameRounding, 5).
		SetStyleFloat(giu.StyleVarFrameBorderSize, 0).
		SetColor(giu.StyleColorButton, col).
		SetColor(giu.StyleColorButtonActive, col).
		SetColor(giu.StyleColorButtonHovered, col)
}

func (iv *ivycel) setFonts() {
	// adding more than one default font will merge the two fonts together. the order is important
	// because we want the normal alphanumeric glyphs from HackRegular and not FontAwesome
	giu.Context.FontAtlas.SetDefaultFontFromBytes(fonts.FontAwesome, fonts.NormalFontSize)
	giu.Context.FontAtlas.SetDefaultFontFromBytes(fonts.Hack_Regular, fonts.NormalFontSize)

	iv.boldFont = giu.Context.FontAtlas.AddFontFromBytes("Hack-Bold", fonts.Hack_Bold, fonts.NormalFontSize)
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

	iv.setFonts()
	iv.setStyling()

	wnd.Run(iv.layout)
}
