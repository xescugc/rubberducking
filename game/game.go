package game

import (
	"bytes"
	"image"
	"log"
	"time"

	"github.com/ebitenui/ebitenui"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/solarlune/resolv"
	"github.com/xescugc/go-flux/v2"
	"github.com/xescugc/rubberducking/assets"
)

var (
	duckImg         image.Image
	speechBallonImg image.Image
)

func init() {
	di, _, err := image.Decode(bytes.NewReader(assets.Duck_PNG))
	if err != nil {
		log.Fatal(err)
	}
	duckImg = ebiten.NewImageFromImage(ebiten.NewImageFromImage(di).SubImage(image.Rect(0, 0, 16, 16)))

	sbi, _, err := image.Decode(bytes.NewReader(assets.SpeechBallon_PNG))
	if err != nil {
		log.Fatal(err)
	}
	speechBallonImg = ebiten.NewImageFromImage(sbi)
}

type Game struct {
	Store *Store
	AD    *ActionDispatcher

	ui *ebitenui.UI

	speechBallonW   *widget.Window
	speechBallonRC  *widget.Container
	speechBallonTxt *widget.Text
}

func NewGame(d *flux.Dispatcher[*Action], port string, verbose bool) *Game {
	g := &Game{
		Store: NewStore(d, time.Second*10, time.Second*15),
		AD:    NewActionDispatcher(d),
	}

	go g.startHttpServer(port, verbose)

	g.buildUI()

	g.AD.AddMessage("Quack!")

	return g
}

func (g *Game) Update() error {
	state := g.Store.GetState()

	if !ebiten.IsWindowMousePassthrough() {
		ebiten.SetWindowMousePassthrough(true)
	}

	mx, my := ebiten.CursorPosition()
	mr := resolv.NewRectangle(float64(mx), float64(my), 1, 1)
	if mr.IsContainedBy(state.Avatar) {
		if ebiten.IsWindowMousePassthrough() {
			ebiten.SetWindowMousePassthrough(false)
		}
	} else {
		if !ebiten.IsWindowMousePassthrough() {
			ebiten.SetWindowMousePassthrough(true)
		}
	}
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonMiddle) {
		if mr.IsContainedBy(state.Avatar) {
			g.AD.DragAvatar(mx, my)
		}
	}

	g.ui.Update()
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	state := g.Store.GetState()

	g.AD.TPS()

	if time.Now().Sub(state.WokeUpAt) > state.WokeUpTimouet {
		return
	}

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(state.Scale, state.Scale)
	op.GeoM.Translate(state.Avatar.Bounds().Min.X, state.Avatar.Bounds().Min.Y)
	screen.DrawImage(duckImg.(*ebiten.Image), op)

	if state.Message != "" {
		g.speechBallonTxt.Label = state.Message
		g.speechBallonRC.BackgroundImage = LoadImageNineSlice(ScaleImage(speechBallonImg.(*ebiten.Image), int(state.Scale)), 1, 1)
		b := state.Avatar.Bounds()

		//Get the preferred size of the content
		x, y := g.speechBallonW.Contents.PreferredSize()
		//Create a rect with the preferred size of the content
		r := image.Rect(0, 0, x, y)
		//Use the Add method to move the window to the specified point
		r = r.Add(image.Point{int(b.Min.X), int(b.Min.Y) - y})
		//Set the windows location to the rect.
		g.speechBallonW.SetLocation(r)
		//Add the window to the UI.
		//Note: If the window is already added, this will just move the window and not add a duplicate.
		g.ui.AddWindow(g.speechBallonW)
	} else {
		if g.ui.IsWindowOpen(g.speechBallonW) {
			g.speechBallonW.Close()
		}
	}

	g.ui.Draw(screen)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return outsideWidth, outsideHeight
}

func (g *Game) buildUI() {
	rootContainer := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewStackedLayout()),
	)
	ui := &ebitenui.UI{
		Container: rootContainer,
	}

	g.ui = ui

	speechBallonAC := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
	)
	speechBallonRC := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Padding(
				widget.Insets{
					Top:    20,
					Bottom: 40,
					Right:  20,
					Left:   20,
				},
			),
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
		)),
		widget.ContainerOpts.BackgroundImage(LoadImageNineSlice(speechBallonImg.(*ebiten.Image), 1, 1)),
		widget.ContainerOpts.WidgetOpts(),
	)
	speechBallonTxt := widget.NewText(
		widget.TextOpts.Text("", Font20, Black),
		widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionStart),
		widget.TextOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionStart,
			}),
		),
	)
	speechBallonRC.AddChild(speechBallonTxt)
	speechBallonAC.AddChild(speechBallonRC)

	speechBallonW := widget.NewWindow(
		//Set the main contents of the window
		widget.WindowOpts.Contents(speechBallonAC),
		//Set the window above everything else and block input elsewhere
		widget.WindowOpts.Modal(),
	)

	g.speechBallonW = speechBallonW
	g.speechBallonTxt = speechBallonTxt
	g.speechBallonRC = speechBallonRC
}
