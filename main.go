package main

import (
	"fmt"
	"image"
	"image/color"
	"strings"

	imgui "github.com/AllenDang/cimgui-go"
	"github.com/AllenDang/giu"
	"github.com/jetsetilly/ivycel/cells"
	"github.com/jetsetilly/ivycel/engine"
	"github.com/jetsetilly/ivycel/engine/ivy"
	"github.com/jetsetilly/ivycel/fonts"
	"github.com/jetsetilly/ivycel/references"
	"github.com/jetsetilly/ivycel/worksheet"
)

type ivycel struct {
	ivy ivy.Ivy

	worksheet worksheet.Worksheet

	cellNormalStyle   *giu.StyleSetter
	cellReadOnlyStyle *giu.StyleSetter
	cellEditStyle     *giu.StyleSetter
	contextMenuStyle  *giu.StyleSetter
	headerStyle       *giu.StyleSetter

	// badge styling should push the badges style first and then the specific badge type
	badges          *giu.StyleSetter
	outputBaseBadge *giu.StyleSetter
	inputBaseBadge  *giu.StyleSetter

	statusBarHeight int

	boldFont *giu.FontInfo

	// fonts should be prepared as soon as possible in order for them to
	// be ready on the frame they are required. without preloading the
	// loading may be visible to the user
	fontsPreloaded bool

	// the current context menu level being shown for cells
	cellContextMenuLevel int
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

// insert text into to cell being edited. insertion is done at the current
// cursor position
func (iv *ivycel) insertIntoCellEdit(insert string) {
	editCell := iv.worksheet.User.(*worksheetUser).editing
	pos := editCell.User.(*cellUser).editCursorPosition

	editCell.Entry = fmt.Sprintf("%s%s%s",
		editCell.Entry[:pos],
		insert,
		editCell.Entry[pos:],
	)

	// advance cursor position of edit cell
	editCell.User.(*cellUser).editCursorPosition += len(insert)

	// cell being editing will need to be refocused after double-click
	iv.worksheet.User.(*worksheetUser).focusCell = true
}

// cell context menu is drawn for cell but not if it's being edited. however, if another cell is
// being edited then that will affect the options offered.
func (iv *ivycel) cellContextMenu(cell *cells.Cell) giu.Widget {
	if iv.worksheet.User.(*worksheetUser).editing != nil {
		editCell := iv.worksheet.User.(*worksheetUser).editing
		menu := giu.ContextMenu().Layout(
			giu.Custom(func() {
				iv.contextMenuStyle.Push()
				defer iv.contextMenuStyle.Pop()
				giu.Column(
					giu.Label(fmt.Sprintf("Editing %s", editCell.Position().Reference())),
					giu.Spacing(),
					giu.Separator(),
					giu.Spacing(),
				).Build()

				giu.Label("Insert...").Build()

				giu.Selectable(fmt.Sprintf(" Reference to %s", cell.Position().Reference())).
					OnClick(func() {
						iv.insertIntoCellEdit(references.WrapCellReference(cell.Position().Reference()))
					}).
					Build()

				if cell.Parent() != nil {
					giu.Selectable(fmt.Sprintf(" Reference to parent (%s)", cell.Parent().Position().Reference())).
						OnClick(func() {
							iv.insertIntoCellEdit(references.WrapCellReference(cell.Parent().Position().Reference()))
						}).
						Build()

				} else if cell.HasChildren() {
					giu.Selectable(
						fmt.Sprintf(" Reference to root of %s", cell.Position().Reference())).
						OnClick(func() {
							iv.insertIntoCellEdit(references.WrapCellReference(
								fmt.Sprintf("%s%s", cell.Position().Reference(), cell.RootIndex()),
							))
						}).
						Build()
				}

				if strings.TrimSpace(cell.Result()) != "" {
					giu.Selectable(fmt.Sprintf(" Literal value of %v", cell.Result())).
						OnClick(func() {
							iv.insertIntoCellEdit(cell.Result())
						}).
						Build()
				}
			}),
		)
		return menu
	}

	cellBase := cell.Base()

	inputBase := func(label string, newBase int) giu.Widget {
		return giu.MenuItem(label).Selected(cellBase.Input == newBase).OnClick(func() {
			cell.SetBase(engine.Base{Input: newBase, Output: cellBase.Output})
			iv.worksheet.RecalculateAll()
		})
	}

	outputBase := func(label string, newBase int) giu.Widget {
		return giu.MenuItem(label).Selected(cellBase.Output == newBase).OnClick(func() {
			cell.SetBase(engine.Base{Input: cellBase.Input, Output: newBase})
			iv.worksheet.RecalculateAll()
		})
	}

	return giu.ContextMenu().Layout(
		giu.Custom(func() {
			iv.contextMenuStyle.Push()
			defer iv.contextMenuStyle.Pop()

			switch iv.cellContextMenuLevel {
			case 0:
				giu.Column(
					giu.Label(fmt.Sprintf("Cell %s", cell.Position().Reference())),

					giu.Spacing(),
					giu.Separator(),
					giu.Spacing(),

					giu.Custom(func() {
						sty := giu.Style().SetDisabled(cell.Entry == "")
						sty.Push()
						defer sty.Pop()
						giu.Selectable("Clear").OnClick(func() {
							cell.Entry = ""
							cell.Commit(true)
							iv.worksheet.RecalculateAll()
						}).Build()
					}),

					giu.Selectable("Input Base...").
						Flags(giu.SelectableFlagsDontClosePopups).
						OnClick(func() {
							iv.cellContextMenuLevel = 1
						}),

					giu.Selectable("Output Base...").
						Flags(giu.SelectableFlagsDontClosePopups).
						OnClick(func() {
							iv.cellContextMenuLevel = 2
						}),

					giu.Custom(func() {
						sty := giu.Style().SetDisabled(cellBase == iv.ivy.Base())
						sty.Push()
						defer sty.Pop()
						giu.Selectable("Reset Bases").OnClick(func() {
							cell.SetBase(iv.ivy.Base())
							iv.worksheet.RecalculateAll()
						}).Build()
					}),
				).Build()
			case 1:
				giu.Column(
					giu.Row(
						giu.Button(string(fonts.ParentContextMenu)).OnClick(func() {
							iv.cellContextMenuLevel = 0
						}),
						giu.Label(fmt.Sprintf("Input Base for %s", cell.Position().Reference())),
					),
					giu.Spacing(),
					giu.Separator(),
					giu.Spacing(),
					inputBase("Binary", 2),
					inputBase("Octal", 8),
					inputBase("Decimal", 10),
					inputBase("Hexadecimal", 16),
				).Build()
			case 2:
				giu.Column(
					giu.Row(
						giu.Button(string(fonts.ParentContextMenu)).OnClick(func() {
							iv.cellContextMenuLevel = 0
						}),
						giu.Label(fmt.Sprintf("Output Base for %s", cell.Position().Reference())),
					),
					giu.Spacing(),
					giu.Separator(),
					giu.Spacing(),
					outputBase("Binary", 2),
					outputBase("Octal", 8),
					outputBase("Decimal", 10),
					outputBase("Hexadecimal", 16),
				).Build()
			}
		}),
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
		rowCount, colCt := iv.worksheet.Size()

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
			for coli := range colCt {
				rowCols = append(rowCols,
					giu.Custom(func() {
						col := cells.NumericToBase26(coli)
						iv.headerStyle.To(
							giu.Button(col).
								Size(-1, fonts.WorksheetHeaderSize),
						).Build()
						giu.ContextMenu().Layout(giu.Custom(func() {
							iv.contextMenuStyle.Push()
							defer iv.contextMenuStyle.Pop()
							giu.Column(
								giu.Selectable(fmt.Sprintf("Insert column before %s", col)).
									OnClick(func() {
										iv.worksheet.InsertColumn(coli)
									}),
							).Build()
						})).Build()
					}),
				)
			}
			rows = append(rows, giu.TableRow(rowCols...))
		}

		for rowi := range rowCount {
			var rowCols []giu.Widget

			// first column of each row is the row number
			rowCols = append(rowCols, giu.Custom(func() {
				lbl := fmt.Sprintf(" %d", rowi+1)
				w, _ := giu.CalcTextSize(lbl)
				iv.headerStyle.To(
					giu.Button(lbl).Size(w, rowHeight),
				).Build()
				giu.ContextMenu().Layout(giu.Custom(func() {
					iv.contextMenuStyle.Push()
					defer iv.contextMenuStyle.Pop()
					giu.Column(
						giu.Selectable(fmt.Sprintf("Insert row before row %d", rowi+1)).
							OnClick(func() {
								iv.worksheet.InsertRow(rowi)
							}),
					).Build()
				})).Build()
			}))

			for coli := range colCt {
				// reference to the cell at row/column number
				cell := iv.worksheet.Cell(rowi, coli)

				var badges giu.Widget
				if !cell.ReadOnly() {
					bs := cell.Base()
					badges = giu.Custom(func() {
						var pos image.Point
						giu.SameLine()
						pos = giu.GetCursorScreenPos().Sub(image.Point{X: 5, Y: 2})

						const badgeSpacing = 8

						iv.badges.Push()
						defer iv.badges.Pop()

						if bs.Output != iv.ivy.Base().Output {
							iv.outputBaseBadge.Push()
							defer iv.outputBaseBadge.Pop()
							txt := fmt.Sprintf("%d", bs.Output)
							pos = pos.Sub(image.Point{X: int(imgui.CalcTextSize(txt).X) + badgeSpacing})
							giu.SetCursorScreenPos(pos)
							giu.Button(txt).Build()
						}

						if bs.Input != iv.ivy.Base().Input {
							iv.inputBaseBadge.Push()
							defer iv.inputBaseBadge.Pop()
							txt := fmt.Sprintf("%d", bs.Input)
							pos = pos.Sub(image.Point{X: int(imgui.CalcTextSize(txt).X) + badgeSpacing})
							giu.SetCursorScreenPos(pos)
							giu.Button(txt).Build()
						}
					})
				}

				// how we display the cell depends on whether the cell is the
				// one currently being edited
				if iv.worksheet.User.(*worksheetUser).editing == cell {
					giu.SetKeyboardFocusHere()
					celInp := giu.InputText(&cell.Entry).Size(-1)

					// escape key cancels changes and deactivates the input text
					// for the cell
					if giu.IsKeyPressed(giu.KeyEscape) {
						iv.worksheet.User.(*worksheetUser).editing = nil
					}

					// CalbackAlways flag so we can update the editCursorPosition every keypress
					// and EnterReturnsTrue so that OnChange() is not triggered until editing
					// has finished
					celInp.Flags(giu.InputTextFlagsCallbackAlways | giu.InputTextFlagsEnterReturnsTrue)

					// keep track of current cursor position in the input
					// widget. we use this to insert cell references at the
					// correct point
					celInp.Callback(func(data imgui.InputTextCallbackData) int {
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
					celInp.OnChange(func() {
						iv.worksheet.User.(*worksheetUser).editing = nil
						cell.Commit(true)
						iv.worksheet.RecalculateAll()
					})

					rowCols = append(rowCols,
						giu.Custom(func() {
							iv.cellEditStyle.Push()
							defer iv.cellEditStyle.Pop()
							if iv.worksheet.User.(*worksheetUser).focusCell {
								// focusCell flag will be reset in the input widget's callback
								// function above. we do this because setting the keyboard focus
								// selects the entire contents of the input and we don't want that.
								// delaying the flag reset allows us to clear the selection and move
								// the input cursor
								giu.SetKeyboardFocusHere()
							}
							celInp.Build()
							if badges != nil {
								badges.Build()
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
							iv.insertIntoCellEdit(references.WrapCellReference(cell.Position().Reference()))
						} else if !cell.ReadOnly() {
							iv.worksheet.User.(*worksheetUser).editing = cell
							iv.worksheet.User.(*worksheetUser).focusCell = true
						}
					})

					// decide on display style for cell
					var sty *giu.StyleSetter
					if cell.ReadOnly() {
						sty = iv.cellReadOnlyStyle
					} else {
						sty = iv.cellNormalStyle

					}

					rowCols = append(rowCols,
						giu.Custom(func() {
							sty.Push()
							defer sty.Pop()
							giu.Row(
								cel, iv.cellContextMenu(cell),
								ev, tip,
							).Build()
							if badges != nil {
								badges.Build()
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
		giu.MenuBar().Layout(
			giu.Spacing(),
			giu.Menu(string(fonts.FileMenu)).Layout(
				giu.Label("File"),
				giu.Separator(),
				giu.MenuItem("Open..."),
				giu.MenuItem("Save..."),
			),
		),
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
	iv.cellNormalStyle = giu.Style().
		SetStyleFloat(giu.StyleVarFrameBorderSize, 0).
		SetStyleFloat(giu.StyleVarFrameRounding, 0).
		SetStyle(giu.StyleVarButtonTextAlign, 0, 0).
		SetColor(giu.StyleColorText, color.RGBA{R: 255, G: 255, B: 255, A: 255})

	iv.cellReadOnlyStyle = giu.Style().
		SetStyleFloat(giu.StyleVarFrameBorderSize, 0).
		SetStyleFloat(giu.StyleVarFrameRounding, 0).
		SetStyle(giu.StyleVarButtonTextAlign, 0, 0).
		SetColor(giu.StyleColorButton, color.Transparent)

	iv.cellEditStyle = giu.Style().
		SetStyleFloat(giu.StyleVarFrameBorderSize, 2).
		SetStyleFloat(giu.StyleVarFrameRounding, 3).
		SetColor(giu.StyleColorBorder, color.RGBA{R: 100, G: 100, B: 200, A: 255})

	iv.contextMenuStyle = giu.Style().
		SetFontSize(fonts.ContextMenuFontSize)

	vcol := imgui.CurrentStyle().Colors()[giu.StyleColorTableRowBg]
	col := color.RGBA{R: uint8(vcol.X), G: uint8(vcol.Y), B: uint8(vcol.Z), A: uint8(vcol.W)}

	iv.headerStyle = giu.Style().
		SetFontSize(fonts.WorksheetHeaderSize).
		SetStyleFloat(giu.StyleVarFrameBorderSize, 0).
		SetStyle(giu.StyleVarFramePadding, 0, 0).
		SetColor(giu.StyleColorButton, col).
		SetColor(giu.StyleColorButtonActive, col).
		SetColor(giu.StyleColorButtonHovered, col)

	iv.badges = giu.Style().
		SetFont(iv.boldFont).
		SetFontSize(fonts.BadgeFontSize).
		SetStyle(giu.StyleVarFramePadding, 2, 2).
		SetStyleFloat(giu.StyleVarFrameRounding, 5).
		SetStyleFloat(giu.StyleVarFrameBorderSize, 0)

	col = color.RGBA{R: 255, G: 100, B: 100, A: 200}
	iv.outputBaseBadge = giu.Style().
		SetColor(giu.StyleColorButton, col).
		SetColor(giu.StyleColorButtonActive, col).
		SetColor(giu.StyleColorButtonHovered, col)

	col = color.RGBA{R: 100, G: 100, B: 255, A: 200}
	iv.inputBaseBadge = giu.Style().
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

	addCellUser := func(cell *cells.Cell) {
		cell.User = &cellUser{}
	}

	iv.worksheet = worksheet.NewWorksheet(&iv.ivy, 30, 30, addCellUser)
	iv.worksheet.User = &worksheetUser{
		selected:     iv.worksheet.Cell(0, 0),
		focusFormula: true,
	}

	wnd := giu.NewMasterWindow("Ivycel", 800, 600, 0)

	iv.setFonts()
	iv.setStyling()

	wnd.Run(iv.layout)
}
