package assets

import (
	_ "embed"
	_ "image/png"
)

//go:embed duck.png
var Duck_PNG []byte

//go:embed speech-ballon.png
var SpeechBallon_PNG []byte

//go:embed buttons.png
var Buttons_PNG []byte

//go:embed frame-bg.png
var FrameBG_PNG []byte

//go:embed input.png
var Input_PNG []byte

//go:embed left-tab.png
var LeftTab_PNG []byte

//go:embed center-tab.png
var CenterTab_PNG []byte

//go:embed right-tab.png
var RightTab_PNG []byte

//go:embed checkbox.png
var Checkbox_PNG []byte

//go:embed munro.ttf
var Munro_TTF []byte

//go:embed game
var Game []byte
