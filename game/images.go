package main

import (
	"bytes"
	"image"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/xescugc/rubberducking/assets"
)

var (
	Images *ImagesCache
)

type ImageKey int

const (
	NoImageKey ImageKey = iota
	DuckKey
	SpeechBallonKey
	FrameBGKey
	InputKey

	ButtonNormalKey
	ButtonHoverKey
	ButtonPressedKey

	CheckboxNormalKey
	CheckboxPressedKey

	LeftTabNormalKey
	LeftTabPressedKey

	CenterLeftTabNormalKey
	CenterRightTabNormalKey
	CenterTabPressedKey

	RightTabNormalKey
	RightTabPressedKey
)

// ImagesCache is a simple cache for all the images, so instead
// of running 'ebiten.NewImageFromImage' we just ran it once
// and reuse it all the time
type ImagesCache struct {
	images map[ImageKey]*ebiten.Image
}

// Get will return the image from 'key', if it does not
// exists a 'nil' will be returned
func (i *ImagesCache) Get(key ImageKey) *ebiten.Image {
	ei, _ := i.images[key]

	return ei
}

type imageSlice struct {
	img []byte
	rec image.Rectangle
}

func init() {
	Images = &ImagesCache{
		images: make(map[ImageKey]*ebiten.Image),
	}
	var keyImage = map[ImageKey]imageSlice{
		DuckKey:         {img: assets.Duck_PNG},
		SpeechBallonKey: {img: assets.SpeechBallon_PNG},
		FrameBGKey:      {img: assets.FrameBG_PNG},
		InputKey:        {img: assets.Input_PNG},
		ButtonNormalKey: {
			img: assets.Buttons_PNG,
			rec: image.Rect(0, 0, 6, 9),
		},
		ButtonHoverKey: {
			img: assets.Buttons_PNG,
			rec: image.Rect(7, 0, 13, 9),
		},
		ButtonPressedKey: {
			img: assets.Buttons_PNG,
			rec: image.Rect(14, 0, 20, 9),
		},
		CheckboxNormalKey: {
			img: assets.Checkbox_PNG,
			rec: image.Rect(0, 0, 9, 7),
		},
		CheckboxPressedKey: {
			img: assets.Checkbox_PNG,
			rec: image.Rect(9, 0, 18, 7),
		},

		LeftTabNormalKey: {
			img: assets.LeftTab_PNG,
			rec: image.Rect(0, 0, 5, 5),
		},
		LeftTabPressedKey: {
			img: assets.LeftTab_PNG,
			rec: image.Rect(5, 0, 10, 5),
		},

		CenterLeftTabNormalKey: {
			img: assets.CenterTab_PNG,
			rec: image.Rect(0, 0, 5, 5),
		},
		CenterRightTabNormalKey: {
			img: assets.CenterTab_PNG,
			rec: image.Rect(5, 0, 10, 5),
		},
		CenterTabPressedKey: {
			img: assets.CenterTab_PNG,
			rec: image.Rect(10, 0, 15, 5),
		},

		RightTabNormalKey: {
			img: assets.RightTab_PNG,
			rec: image.Rect(0, 0, 5, 5),
		},
		RightTabPressedKey: {
			img: assets.RightTab_PNG,
			rec: image.Rect(5, 0, 10, 5),
		},
	}

	var nir image.Rectangle
	for k, is := range keyImage {
		i, _, err := image.Decode(bytes.NewReader(is.img))
		if err != nil {
			log.Fatalf("failed to decode image with key %q: %s", k, err)
		}
		if is.rec != nir {
			i = ebiten.NewImageFromImage(i).SubImage(is.rec)
		}
		Images.images[k] = ebiten.NewImageFromImage(i)
	}

}
