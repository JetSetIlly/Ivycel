package fonts

import _ "embed"

//go:embed "HackNerdFont-Regular.ttf"
var HackNerd_Regular []byte

//go:embed "Hack-Bold.ttf"
var Hack_Bold []byte

const (
	FileMenu   = rune(0xf15b)
	InputBase  = rune(0xf090)
	OutputBase = rune(0xf08b)
)
