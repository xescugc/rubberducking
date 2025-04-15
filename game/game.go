package main

import (
	"image"
	"time"

	"github.com/ebitenui/ebitenui"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/solarlune/resolv"
	"github.com/xescugc/go-flux/v2"
)

type Game struct {
	Store *Store
	AD    *ActionDispatcher

	ui *ebitenui.UI

	speechBallonW   *widget.Window
	speechBallonRC  *widget.Container
	speechBallonTxt *widget.Text

	menuW         *widget.Window
	focusCheckbox *widget.Button
}

func NewGame(d *flux.Dispatcher[*Action], s *Store, ad *ActionDispatcher) *Game {
	g := &Game{
		Store: s,
		AD:    ad,
	}

	g.buildUI()

	return g
}

func (g *Game) Update() error {
	state := g.Store.GetState()

	if time.Now().Sub(state.WokeUpAt) > state.WokeUpTimeout && !state.MenuOpen {
		return ebiten.Termination
	}

	if !ebiten.IsWindowMousePassthrough() {
		ebiten.SetWindowMousePassthrough(true)
	}

	mx, my := ebiten.CursorPosition()
	mr := resolv.NewRectangle(float64(mx), float64(my), 1, 1)
	if mr.IsContainedBy(state.Avatar) || state.MenuOpen {
		//if mr.IsContainedBy(state.Avatar) {
		if ebiten.IsWindowMousePassthrough() {
			ebiten.SetWindowMousePassthrough(false)
		}
		//} else if state.MenuOpen {
		//rec := g.menuW.GetContainer().GetWidget().Rect
		//wres := resolv.NewRectangleFromTopLeft(float64(rec.Min.X), float64(rec.Min.Y), float64(rec.Max.X-rec.Min.X), float64(rec.Max.Y-rec.Min.Y))
		//if mr.IsContainedBy(wres) {
		//if ebiten.IsWindowMousePassthrough() {
		//ebiten.SetWindowMousePassthrough(false)
		//}
		//} else {
		//if !ebiten.IsWindowMousePassthrough() {
		//ebiten.SetWindowMousePassthrough(true)
		//}
		//}
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

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight) {
		if mr.IsContainedBy(state.Avatar) {
			g.AD.MenuOpen(true)
		}
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) && state.MenuOpen {
		g.AD.MenuOpen(false)
	}

	g.ui.Update()
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	state := g.Store.GetState()

	g.AD.TPS()

	if _, ok := g.Store.GetMessage(); !ok && !g.Store.GetDisplay() {
		return
	}

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(state.Scale, state.Scale)
	op.GeoM.Translate(state.Avatar.Bounds().Min.X, state.Avatar.Bounds().Min.Y)
	screen.DrawImage(Images.Get(DuckKey), op)

	if m, ok := g.Store.GetMessage(); ok {
		g.speechBallonTxt.Label = m.Text
		g.speechBallonRC.BackgroundImage = LoadImageNineSlice(SpeechBallonKey, int(state.Scale), 1, 1)

		b := state.Avatar.Bounds()
		msw, msh := ebiten.Monitor().Size()

		//Get the preferred size of the content
		x, y := g.speechBallonW.Contents.PreferredSize()

		endx := int(b.Min.X)
		endy := int(b.Min.Y) - y

		if endx+x > msw {
			endx -= (endx + x) - msw
		} else if endx < 0 {
			endx = 0
		}

		if endy+y > msh {
			endy -= (endy + y) - msh
		} else if endy < 0 {
			endy = 0
		}

		//Create a rect with the preferred size of the content
		r := image.Rect(0, 0, x, y)
		//Use the Add method to move the window to the specified point
		r = r.Add(image.Point{endx, endy})
		//Set the windows location to the rect.
		g.speechBallonW.SetLocation(r)
		//Add the window to the UI.
		//Note: If the window is already added, this will just move the window and not add a duplicate.
		g.ui.AddWindow(g.speechBallonW)
	} else {
		if g.ui.IsWindowOpen(g.speechBallonW) {
			g.menuW.Close()
		}
	}

	if state.MenuOpen {
		msw, msh := ebiten.Monitor().Size()

		//Get the preferred size of the content
		x, y := g.menuW.Contents.PreferredSize()

		//Create a rect with the preferred size of the content
		r := image.Rect(0, 0, x, y)
		//Use the Add method to move the window to the specified point
		r = r.Add(image.Point{(msw / 2) - x/2, (msh / 2) - y/2})
		//Set the windows location to the rect.
		g.menuW.SetLocation(r)
		//Add the window to the UI.
		//Note: If the window is already added, this will just move the window and not add a duplicate.
		g.ui.AddWindow(g.menuW)

		if state.FocusMode {
			g.focusCheckbox.SetState(widget.WidgetChecked)
		} else {
			g.focusCheckbox.SetState(widget.WidgetUnchecked)
		}
	} else {
		if g.ui.IsWindowOpen(g.menuW) {
			g.menuW.Close()
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

	g.buildSpeechBallonW()
	g.buildMenuW()
}

func (g *Game) buildSpeechBallonW() {
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
		widget.ContainerOpts.BackgroundImage(LoadImageNineSlice(SpeechBallonKey, 1, 1, 1)),
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

func (g *Game) buildMenuW() {
	menuAC := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
	)
	menuRC := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Padding(
				widget.Insets{
					Top:    5,
					Bottom: 40,
					Right:  5,
					Left:   5,
				},
			),
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
		)),
		widget.ContainerOpts.BackgroundImage(LoadImageNineSlice(FrameBGKey, 5, 1, 1)),
	)
	menuAC.AddChild(menuRC)

	tabsGC := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Columns(3),
			widget.GridLayoutOpts.Spacing(0, 0),
			widget.GridLayoutOpts.DefaultStretch(true, true),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionCenter,
				Stretch:  true,
			}),
		),
	)

	var tab1BodyRC, tab2BodyRC, tab3BodyRC *widget.Container

	tabTextInsets := widget.Insets{
		Left:   20,
		Right:  20,
		Top:    10,
		Bottom: 22,
	}

	tab1Btn := widget.NewButton(
		// specify the images to sue
		widget.ButtonOpts.Image(LeftTabButtonResource()),

		// specify the button's text, the font face, and the color
		widget.ButtonOpts.Text("Actions", Font30, ButtonTextColor()),

		// specify that the button's text needs some padding for correct display
		widget.ButtonOpts.TextPadding(tabTextInsets),
		widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
			tab1BodyRC.GetWidget().Visibility = widget.Visibility_Show
			tab2BodyRC.GetWidget().Visibility = widget.Visibility_Hide
			tab3BodyRC.GetWidget().Visibility = widget.Visibility_Hide
		}),
	)
	tab2Btn := widget.NewButton(
		// specify the images to sue
		widget.ButtonOpts.Image(CenterRightTabButtonResource()),

		// specify the button's text, the font face, and the color
		widget.ButtonOpts.Text("Config", Font30, ButtonTextColor()),

		// specify that the button's text needs some padding for correct display
		widget.ButtonOpts.TextPadding(tabTextInsets),
		widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
			tab1BodyRC.GetWidget().Visibility = widget.Visibility_Hide
			tab2BodyRC.GetWidget().Visibility = widget.Visibility_Show
			tab3BodyRC.GetWidget().Visibility = widget.Visibility_Hide
		}),
	)
	tab3Btn := widget.NewButton(
		// specify the images to sue
		widget.ButtonOpts.Image(RightTabButtonResource()),

		// specify the button's text, the font face, and the color
		widget.ButtonOpts.Text("About", Font30, ButtonTextColor()),

		// specify that the button's text needs some padding for correct display
		widget.ButtonOpts.TextPadding(tabTextInsets),
		widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
			tab1BodyRC.GetWidget().Visibility = widget.Visibility_Hide
			tab2BodyRC.GetWidget().Visibility = widget.Visibility_Hide
			tab3BodyRC.GetWidget().Visibility = widget.Visibility_Show
		}),
	)
	tabsGC.AddChild(
		tab1Btn,
		tab2Btn,
		tab3Btn,
	)

	widget.NewRadioGroup(
		widget.RadioGroupOpts.Elements(
			tab1Btn,
			tab2Btn,
			tab3Btn,
		),
	)
	tabBodyInsets := widget.Insets{
		Left:   25,
		Right:  25,
		Top:    25,
		Bottom: 25,
	}

	tab1BodyRC = widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Columns(2),
			widget.GridLayoutOpts.Spacing(0, 0),
			widget.GridLayoutOpts.Stretch([]bool{true, false}, []bool{false}),
			widget.GridLayoutOpts.Padding(tabBodyInsets),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionCenter,
				Stretch:  true,
			}),
		),
	)

	focusLabel := widget.NewText(
		widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionCenter),
		widget.TextOpts.Text("Focus mode", Font20, Black),
	)
	focusCheckbox := widget.NewButton(
		// specify the images to sue
		widget.ButtonOpts.Image(CheckboxButtonResource()),
		widget.ButtonOpts.ToggleMode(),
		widget.ButtonOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.GridLayoutData{
				MaxWidth:  45,
				MaxHeight: 35,
			}),
		),
		widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
			if args.Button.State() == widget.WidgetUnchecked {
				g.AD.SetFocusMode(false)
			} else {
				g.AD.SetFocusMode(true)
			}
		}),
	)

	g.focusCheckbox = focusCheckbox

	tab1BodyRC.AddChild(
		focusLabel, focusCheckbox,
	)

	tab2BodyRC = widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Columns(2),
			widget.GridLayoutOpts.Spacing(0, 0),
			widget.GridLayoutOpts.Stretch([]bool{true, false}, []bool{false}),
			widget.GridLayoutOpts.Padding(tabBodyInsets),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionCenter,
				Stretch:  true,
			}),
			func(w *widget.Widget) {
				w.Visibility = widget.Visibility_Hide
			},
		),
	)
	tab2Txt := widget.NewText(
		widget.TextOpts.Text("Config", Font20, Black),
		widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionStart),
	)
	tab2BodyRC.AddChild(tab2Txt)

	tab3BodyRC = widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Columns(2),
			widget.GridLayoutOpts.Spacing(0, 0),
			widget.GridLayoutOpts.Stretch([]bool{true, false}, []bool{false}),
			widget.GridLayoutOpts.Padding(tabBodyInsets),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionCenter,
				Stretch:  true,
			}),
			func(w *widget.Widget) {
				w.Visibility = widget.Visibility_Hide
			},
		),
	)
	tab3Txt := widget.NewText(
		widget.TextOpts.Text("About", Font20, Black),
		widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionStart),
	)
	tab3BodyRC.AddChild(tab3Txt)

	menuRC.AddChild(
		tabsGC,
		tab1BodyRC,
		tab2BodyRC,
		tab3BodyRC,
	)

	menuW := widget.NewWindow(
		widget.WindowOpts.Contents(menuAC),
		widget.WindowOpts.Modal(),
		widget.WindowOpts.CloseMode(widget.CLICK_OUT),

		widget.WindowOpts.ClosedHandler(func(args *widget.WindowClosedEventArgs) {
			g.AD.MenuOpen(false)
		}),
	)

	g.menuW = menuW
}
