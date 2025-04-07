package main

import (
	"encoding/json"
	"os"
	"sync"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/solarlune/resolv"
	"github.com/spf13/afero"
	"github.com/xescugc/go-flux/v2"
)

type Store struct {
	*flux.ReduceStore[State, *Action]

	mxStore sync.RWMutex
}

type State struct {
	Avatar *resolv.ConvexPolygon `json:"-"`

	Message          string        `json:"-""`
	MessageCreatedAt time.Time     `json:"-"`
	MessageTimeout   time.Duration `json:"message_timout"`

	WokeUpAt      time.Time     `json:"-"`
	WokeUpTimeout time.Duration `json:"woke_up_timeout"`

	Scale float64 `json:"scale"`
}

type AvatarJSON struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	W float64 `json:"w"`
	H float64 `json:"h"`
}

func (s *State) UnmarshalJSON(data []byte) error {
	type Alias State
	aux := &struct {
		Avatar AvatarJSON `json:"avatar"`
		*Alias
	}{
		Alias: (*Alias)(s),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	s.Avatar = resolv.NewRectangleFromTopLeft(aux.Avatar.X, aux.Avatar.Y, aux.Avatar.W, aux.Avatar.H)

	return nil
}

func (s *State) MarshalJSON() ([]byte, error) {
	type Alias State

	ab := s.Avatar.Bounds()
	ms := &struct {
		*Alias
		Avatar AvatarJSON `json:"avatar"`
	}{
		Avatar: AvatarJSON{
			X: ab.Min.X,
			Y: ab.Min.Y,
			W: ab.Max.X - ab.Min.X,
			H: ab.Max.Y - ab.Min.Y,
		},
		Alias: (*Alias)(s),
	}
	Logger.Info("Before sending", "struct", ms)
	return json.Marshal(ms)
}

func NewStore(d *flux.Dispatcher[*Action], fs afero.Fs, mto, wuto time.Duration) *Store {
	s := &Store{}

	initLog()
	state := initialState(fs)
	if state == nil {
		Logger.Info("No state file found, initializing a new one")
		msw, msh := ebiten.Monitor().Size()
		isw, ish := duckImg.(*ebiten.Image).Size()

		scale := 10
		a := resolv.NewRectangleFromTopLeft(
			float64(msw-(isw*scale)),
			float64(msh-(ish*scale)),
			float64(isw*scale),
			float64(ish*scale),
		)

		state = &State{
			Avatar: a,

			Scale: float64(scale),

			MessageTimeout: mto,

			WokeUpAt:      time.Now(),
			WokeUpTimeout: wuto,
		}
	}
	// NOTE: We are not using SetScale as it does not scale correctly
	// in the render side
	//a.SetScale(scale, scale)

	s.ReduceStore = flux.NewReduceStore(d, s.Reduce, *state)

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

func initialState(fs afero.Fs) *State {
	fi, err := fs.Open(stateFile)
	if err != nil {
		if os.IsNotExist(err) {
			Logger.Info("Game state file is not present")
			return nil
		}
		Logger.Error("Error when reading state file", "error", err)
		return nil
	}
	defer fi.Close()

	var state State
	err = json.NewDecoder(fi).Decode(&state)
	if err != nil {
		Logger.Error("Error when decoding state file", "error", err)
		return nil
	}

	return &state
}
