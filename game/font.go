package game

import (
	"bytes"
	"image/color"
	"log"

	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/xescugc/rubberducking/assets"
)

var (
	Font20 text.Face

	Black = color.RGBA{46, 34, 47, 255}
)

func init() {
	mtt, err := text.NewGoTextFaceSource(bytes.NewReader(assets.Munro_TTF))
	if err != nil {
		log.Fatal(err)
	}

	const dpi = 72
	Font20 = &text.GoTextFace{
		Source: mtt,
		Size:   20,
	}
}
