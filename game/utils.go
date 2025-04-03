package game

import (
	"github.com/ebitenui/ebitenui/image"
	"github.com/hajimehoshi/ebiten/v2"
)

func LoadImageNineSlice(i *ebiten.Image, centerWidth, centerHeight int) *image.NineSlice {
	w := i.Bounds().Dx()
	h := i.Bounds().Dy()
	p := 0
	// This means to do it 3x3 equally
	if centerWidth == 0 && centerHeight == 0 {
		centerWidth = w / 3
		centerHeight = h / 3
	}
	return image.NewNineSlice(i,
		[3]int{(w - centerWidth) / 2, centerWidth, (w - centerWidth) / 2},
		[3]int{(h-centerHeight)/2 + p, centerHeight, (h-centerHeight)/2 - p},
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
