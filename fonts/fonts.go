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
	FileMenu   = rune(0xf15b)
	InputBase  = rune(0xf090)
	OutputBase = rune(0xf08b)
)

const (
	NormalFontSize    = 16
	WorksheetFontSize = 18
	BadgeFontSize     = 13
)
