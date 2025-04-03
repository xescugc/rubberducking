package game

import (
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/solarlune/resolv"
	"github.com/xescugc/go-flux/v2"
)

type Store struct {
	*flux.ReduceStore[State, *Action]

	mxLobbies sync.RWMutex
}

type State struct {
	Avatar *resolv.ConvexPolygon

	Message string

	Scale float64
}

func NewStore(d *flux.Dispatcher[*Action]) *Store {
	s := &Store{}

	msw, msh := ebiten.Monitor().Size()
	isw, ish := duckImg.(*ebiten.Image).Size()

	scale := 10
	a := resolv.NewRectangleFromTopLeft(
		float64(msw-(isw*scale)),
		float64(msh-(ish*scale)),
		float64(isw*scale),
		float64(ish*scale),
	)
	// NOTE: We are not using SetScale as it does not scale correctly
	// in the render side
	//a.SetScale(scale, scale)
	s.ReduceStore = flux.NewReduceStore(d, s.Reduce, State{
		Avatar: a,
		Scale:  float64(scale),
	})

	return s
}

func (s *Store) Reduce(state State, act *Action) State {
	switch act.Type {
	case DragAvatar:
		state.Avatar.SetPosition(float64(act.DragAvatar.X), float64(act.DragAvatar.Y))
	case AddMessage:
		state.Message = act.AddMessage.Message
	}

	return state
}
