package assets

import (
	_ "embed"
	_ "image/png"
)

//go:embed duck.png
var Duck_PNG []byte

//go:embed speech-ballon.png
var SpeechBallon_PNG []byte

//go:embed munro.ttf
var Munro_TTF []byte

//go:embed game
var Game []byte
