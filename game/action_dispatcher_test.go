package main

import (
	"testing"
	"time"

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
			MessageTimeout: messageTimeout,
			WokeUpTimouet:  wokeUpTimouet,
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

		require.Equal(t, time.Time{}, as.GetState().MessageCreatedAt)

		ad.AddMessage(msg)

		require.NotEqual(t, time.Time{}, as.GetState().MessageCreatedAt)
		require.NotEqual(t, time.Time{}, as.GetState().WokeUpAt)

		es := testState()
		es.Message = msg
		es.MessageCreatedAt = as.GetState().MessageCreatedAt
		es.WokeUpAt = as.GetState().WokeUpAt

		EqualState(t, es, as.GetState())
	})
}

func TestTPS(t *testing.T) {
	t.Run("ExpireMessage", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		msg := "Quack!"

		ad, as := initStore()
		ad.AddMessage(msg)

		time.Sleep(messageTimeout)

		ad.TPS()

		es := testState()
		es.WokeUpAt = as.GetState().WokeUpAt

		EqualState(t, es, as.GetState())
	})
}

func initStore() (*ActionDispatcher, *Store) {
	d := flux.NewDispatcher[*Action]()
	s := NewStore(d, messageTimeout, wokeUpTimouet)
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
