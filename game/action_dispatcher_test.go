package main

import (
	"testing"
	"time"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xescugc/go-flux/v2"
	"go.uber.org/mock/gomock"
)

var (
	messageTimeout = time.Millisecond
	wokeUpTimouet  = time.Millisecond * 2

	testState = func() State {
		return State{
			MessageTimeout:           messageTimeout,
			WokeUpTimeout:            wokeUpTimouet,
			Messages:                 make([]Message, 0, 0),
			MessageMaxLineCharacters: DefaultMessageMaxLineCharacters,
			MessageMaxLines:          DefaultMessageMaxLines,
		}
	}
)

func TestEmpty(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	_, as := initStore()
	require.NotEqual(t, time.Time{}, as.GetState().WokeUpAt)
	es := testState()
	es.WokeUpAt = as.GetState().WokeUpAt
	EqualState(t, es, as.GetState())
}

func TestDragAvatar(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ad, as := initStore()
		auxA := *as.GetState().Avatar

		ad.DragAvatar(1, 2)

		es := testState()
		auxA.SetPosition(1., 2.)
		es.Avatar = &auxA
		es.WokeUpAt = as.GetState().WokeUpAt

		EqualState(t, es, as.GetState())
	})
}

func TestAddMessage(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		msg := "Quack!"

		ad, as := initStore()

		require.Len(t, as.GetState().Messages, 0)

		ad.AddMessage(msg)

		require.Len(t, as.GetState().Messages, 1)
		require.NotEqual(t, time.Time{}, as.GetState().MessageDisplayedAt)
		require.NotEqual(t, time.Time{}, as.GetState().WokeUpAt)

		es := testState()
		es.Messages = []Message{
			Message{
				Text: msg,
			},
		}
		es.MessageDisplayedAt = as.GetState().MessageDisplayedAt
		es.WokeUpAt = as.GetState().WokeUpAt

		EqualState(t, es, as.GetState())

		t.Run("AddAnother", func(t *testing.T) {
			msg2 := "Quack!2"
			ad.AddMessage(msg2)
			es.Messages = []Message{
				Message{
					Text: msg,
				},
				Message{
					Text: msg2,
				},
			}
			EqualState(t, es, as.GetState())
		})
	})
}

func TestMenuOpen(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ad, as := initStore()

		ad.MenuOpen(true)

		es := testState()
		es.WokeUpAt = as.GetState().WokeUpAt
		es.MenuOpen = true

		EqualState(t, es, as.GetState())
	})
}

func TestSetFocus(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		t.Run("WithNoMessages", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			ad, as := initStore()

			ad.SetFocusMode(true)

			es := testState()
			es.WokeUpAt = as.GetState().WokeUpAt
			es.FocusMode = true

			EqualState(t, es, as.GetState())
		})
		t.Run("WithMessages", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			var (
				msg1 = "msg1"
				msg2 = "msg2"
			)

			ad, as := initStore()

			ad.AddMessage(msg1)
			ad.AddMessage(msg2)
			ad.SetFocusMode(true)

			es := testState()
			es.WokeUpAt = as.GetState().WokeUpAt
			es.FocusMode = true
			es.Messages = []Message{
				Message{
					Text: msg1,
				},
				Message{
					Text: msg2,
				},
			}

			EqualState(t, es, as.GetState())

			t.Run("TPS", func(t *testing.T) {
				ad.TPS()

				EqualState(t, es, as.GetState())

				t.Run("SetItBackTo'False'", func(t *testing.T) {
					ad.SetFocusMode(false)

					es.WokeUpAt = as.GetState().WokeUpAt
					es.MessageDisplayedAt = as.GetState().MessageDisplayedAt
					es.FocusMode = false
					EqualState(t, es, as.GetState())

					t.Run("TPS", func(t *testing.T) {
						time.Sleep(messageTimeout)
						ad.TPS()
						es.WokeUpAt = as.GetState().WokeUpAt
						es.MessageDisplayedAt = as.GetState().MessageDisplayedAt
						es.Messages = []Message{
							Message{
								Text: msg2,
							},
						}
						EqualState(t, es, as.GetState())
					})
				})
			})
		})
	})
}

func TestTPS(t *testing.T) {
	t.Run("ExpireMessage", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		msg := "Quack!"
		msg2 := "Quack!2"

		ad, as := initStore()
		ad.AddMessage(msg)
		ad.AddMessage(msg2)

		time.Sleep(messageTimeout)

		ad.TPS()

		es := testState()

		es.Messages = []Message{
			Message{
				Text: msg2,
			},
		}
		es.MessageDisplayedAt = as.GetState().MessageDisplayedAt
		es.WokeUpAt = as.GetState().WokeUpAt

		EqualState(t, es, as.GetState())
	})
}

func initStore() (*ActionDispatcher, *Store) {
	d := flux.NewDispatcher[*Action]()
	s := NewStore(d, afero.NewMemMapFs(), messageTimeout, wokeUpTimouet)
	return NewActionDispatcher(d), s
}

func EqualState(t *testing.T, e, a State) {
	t.Helper()

	// This values are set during the initialization of the store
	if e.Avatar == nil {
		e.Avatar = a.Avatar
	}
	if e.Scale == 0 {
		e.Scale = a.Scale
	}

	assert.Equal(t, e, a)
}
