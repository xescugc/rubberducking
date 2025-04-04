package game

import (
	"sync"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/solarlune/resolv"
	"github.com/xescugc/go-flux/v2"
)

type Store struct {
	*flux.ReduceStore[State, *Action]

	mxStore sync.RWMutex
}

type State struct {
	Avatar *resolv.ConvexPolygon

	Message          string
	MessageCreatedAt time.Time
	MessageTimeout   time.Duration

	WokeUpAt      time.Time
	WokeUpTimouet time.Duration

	Scale float64
}

func NewStore(d *flux.Dispatcher[*Action], mto, wuto time.Duration) *Store {
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

		Scale: float64(scale),

		MessageTimeout: mto,

		WokeUpTimouet: wuto,
	})

	return s
}

func (s *Store) Reduce(state State, act *Action) State {
	switch act.Type {
	case TPS:
		s.mxStore.Lock()
		defer s.mxStore.Unlock()

		// Remove the message if it has been display long enough
		if state.Message != "" && time.Now().Sub(state.MessageCreatedAt) > state.MessageTimeout {
			state.Message = ""
			state.MessageCreatedAt = time.Time{}
		}
	case DragAvatar:
		s.mxStore.Lock()
		defer s.mxStore.Unlock()

		state.Avatar.SetPosition(float64(act.DragAvatar.X), float64(act.DragAvatar.Y))
	case AddMessage:
		s.mxStore.Lock()
		defer s.mxStore.Unlock()

		state.Message = act.AddMessage.Message
		state.MessageCreatedAt = time.Now()

		state.WokeUpAt = time.Now()
	}

	return state
}
