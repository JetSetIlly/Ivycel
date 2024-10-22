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
)

const (
	NormalFontSize      = 16
	ContextMenuFontSize = 15
	WorksheetFontSize   = 18
	WorksheetHeaderSize = 16
	BadgeFontSize       = 13
)
