package main

import (
	"encoding/json"
	"os"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/solarlune/resolv"
	"github.com/spf13/afero"
	"github.com/xescugc/go-flux/v2"
)

const (
	DefaultMessageMaxLineCharacters = 40
	DefaultMessageMaxLines          = 20
)

type Store struct {
	*flux.ReduceStore[State, *Action]

	mxStore sync.RWMutex
}

type State struct {
	Avatar *resolv.ConvexPolygon `json:"-"`

	Messages           []Message `json:"-"`
	MessageDisplayedAt time.Time `json:"-"`

	MessageTimeout time.Duration `json:"message_timeout"`
	// MessageMaxLineCharacter is the max number characters in one line
	MessageMaxLineCharacters int `json:"message_max_line_characters"`
	// MessageMaxLines is the max number of lines
	MessageMaxLines int `json:"message_max_lines"`

	WokeUpAt      time.Time     `json:"-"`
	WokeUpTimeout time.Duration `json:"woke_up_timeout"`

	// Display is to force displaying even without a message
	Display bool `json:"-"`

	Scale float64 `json:"scale"`
}

type Message struct {
	Text string
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

	state.MessageMaxLineCharacters = DefaultMessageMaxLineCharacters
	state.MessageMaxLines = DefaultMessageMaxLines

	state.Messages = make([]Message, 0, 0)
	// NOTE: We are not using SetScale as it does not scale correctly
	// in the render side
	//a.SetScale(scale, scale)

	s.ReduceStore = flux.NewReduceStore(d, s.Reduce, *state)

	return s
}

func (s *Store) GetMessage() (Message, bool) {
	s.mxStore.Lock()
	defer s.mxStore.Unlock()

	state := s.GetState()
	if len(state.Messages) == 0 {
		return Message{}, false
	}

	return state.Messages[0], true
}

func (s *Store) GetDisplay() bool {
	s.mxStore.Lock()
	defer s.mxStore.Unlock()

	return s.GetState().Display
}

func (s *Store) Reduce(state State, act *Action) State {
	switch act.Type {
	case TPS:
		s.mxStore.Lock()
		defer s.mxStore.Unlock()

		// Remove the message if it has been display long enough
		if len(state.Messages) > 0 && time.Now().Sub(state.MessageDisplayedAt) > state.MessageTimeout {
			state.Messages = state.Messages[1:]
			if len(state.Messages) != 0 {
				state.MessageDisplayedAt = time.Now()
				state.WokeUpAt = time.Now()
			} else {
				state.MessageDisplayedAt = time.Time{}
			}
		}
	case DragAvatar:
		s.mxStore.Lock()
		defer s.mxStore.Unlock()

		state.Avatar.SetPosition(float64(act.DragAvatar.X), float64(act.DragAvatar.Y))
	case AddMessage:
		s.mxStore.Lock()
		defer s.mxStore.Unlock()

		if len(state.Messages) == 0 {
			state.MessageDisplayedAt = time.Now()
			state.WokeUpAt = time.Now()
		}

		state.Messages = append(state.Messages, Message{
			Text: truncateMessage(act.AddMessage.Message, state.MessageMaxLineCharacters, state.MessageMaxLines),
		})
	case Toggle:
		s.mxStore.Lock()
		defer s.mxStore.Unlock()

		if time.Now().Sub(state.WokeUpAt) > state.WokeUpTimeout {
			// If it's sleeping we woke it up
			state.WokeUpAt = time.Now()
			state.Display = true
		} else {
			// If it's woken we make it sleep
			state.WokeUpAt = time.Time{}
			state.Display = false
			state.Messages = nil
		}
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

func truncateMessage(msg string, mlc, ml int) string {
	if len(msg) < mlc {
		return msg
	}
	messages := strings.Split(msg, "\n")
	for i, m := range messages {
		if len(m) < mlc {
			continue
		}

		nms := make([]string, 1, 1)
		ws := strings.Split(m, " ")
		for ii := 0; ii < len(ws); ii++ {
			ci := len(nms) - 1
			cv := nms[ci]
			w := ws[ii]
			// If adding the new w(word) to the
			// line makes it too long
			if len(cv)+len(w) > mlc {
				if cv == "" {
					// This means that is a string longer than the mlc so we have to break it
					for len(w) > mlc {
						ci = len(nms) - 1
						cv = nms[ci]

						// We add '-' at the end so it's clear
						ns := w[0:mlc-1] + "-"
						if len(cv) == mlc {
							nms = append(nms, ns)
						} else {
							nms[ci] = ns
						}

						w = w[mlc-1:]
						if len(w) < mlc {
							nms = append(nms, w)
						}
					}
				} else {
					// We add a new line and move the 'w' to the next index
					// so when it goes again it enters the 'cv==""'
					nms = append(nms, "")
					ws = slices.Insert(ws, i+1, w)
				}
				continue
			}
			if cv == "" {
				nms[ci] = w
			} else {
				nms[ci] = cv + " " + w
			}
		}

		messages[i] = strings.Join(nms, "\n")
	}

	msg = strings.Join(messages, "\n")
	messages = strings.Split(msg, "\n")
	if len(messages) > ml {
		messages = append(messages[:ml], "...")
	}
	return strings.Join(messages, "\n")
}
