package main

import (
	"github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
)

func CheckboxButtonResource() *widget.ButtonImage {
	bn := LoadImageNineSlice(CheckboxNormalKey, 5, 1, 1)
	bp := LoadImageNineSlice(CheckboxPressedKey, 5, 1, 1)

	return &widget.ButtonImage{
		Idle:    bn,
		Pressed: bp,
	}
}

func LeftTabButtonResource() *widget.ButtonImage {
	bn := LoadImageNineSlice(LeftTabNormalKey, 5, 1, 1)
	bp := LoadImageNineSlice(LeftTabPressedKey, 5, 1, 1)

	return &widget.ButtonImage{
		Idle:    bn,
		Pressed: bp,
	}
}

func CenterLeftTabButtonResource() *widget.ButtonImage {
	bn := LoadImageNineSlice(CenterLeftTabNormalKey, 5, 1, 1)
	bp := LoadImageNineSlice(CenterTabPressedKey, 5, 1, 1)

	return &widget.ButtonImage{
		Idle:    bn,
		Pressed: bp,
	}
}

func CenterRightTabButtonResource() *widget.ButtonImage {
	bn := LoadImageNineSlice(CenterRightTabNormalKey, 5, 1, 1)
	bp := LoadImageNineSlice(CenterTabPressedKey, 5, 1, 1)

	return &widget.ButtonImage{
		Idle:    bn,
		Pressed: bp,
	}
}

func RightTabButtonResource() *widget.ButtonImage {
	bn := LoadImageNineSlice(RightTabNormalKey, 5, 1, 1)
	bp := LoadImageNineSlice(RightTabPressedKey, 5, 1, 1)

	return &widget.ButtonImage{
		Idle:    bn,
		Pressed: bp,
	}
}

func ButtonTextColor() *widget.ButtonTextColor {
	return &widget.ButtonTextColor{
		Idle:    White,
		Pressed: Black,
	}
}

func LoadImageNineSlice(ik ImageKey, scale, centerWidth, centerHeight int) *image.NineSlice {
	i := ScaleImage(Images.Get(ik), scale)
	w := i.Bounds().Dx()
	h := i.Bounds().Dy()
	// This means to do it 3x3 equally
	if centerWidth == 0 && centerHeight == 0 {
		centerWidth = w / 3
		centerHeight = h / 3
	}
	return image.NewNineSlice(i,
		[3]int{(w - centerWidth) / 2, centerWidth, (w - centerWidth) / 2},
		[3]int{(h - centerHeight) / 2, centerHeight, (h - centerHeight) / 2},
	)
}

func ScaleImage(i *ebiten.Image, s int) *ebiten.Image {
	sf := float64(s)
	ni := ebiten.NewImage(i.Bounds().Dx()*s, i.Bounds().Dy()*s)

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(sf, sf)
	ni.DrawImage(i, op)

	return ni
}
