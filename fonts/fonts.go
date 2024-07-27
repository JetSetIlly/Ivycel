package fonts

import _ "embed"

//go:embed "Hack-Regular.ttf"
var Hack_Regular []byte

//go:embed "Hack-Bold.ttf"
var Hack_Bold []byte

//go:embed "fa-solid-900.ttf"
var FontAwesome []byte

// Icon glyphs from FontAwesome
const (
	FileMenu = rune(0xf15b)

	// for some reason glyphs from font-awesome do not show in a context menu
	// (although they do in other areas of the gui). therefore, the rune below
	// is a glyph in the hack font
	ParentContextMenu = rune(0x276e)
)

const (
	NormalFontSize      = 16
	ContextMenuFontSize = 15
	WorksheetFontSize   = 18
	BadgeFontSize       = 13
)
