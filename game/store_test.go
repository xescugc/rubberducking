package main

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/solarlune/resolv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStateJSON(t *testing.T) {
	a := resolv.NewRectangleFromTopLeft(
		10, 10,
		5, 5,
	)
	scale := 10
	state := State{
		Avatar: a,

		Scale: float64(scale),

		MessageTimeout: time.Second * 10,

		WokeUpTimeout: time.Second * 15,
	}
	t.Run("Marshal", func(t *testing.T) {
		b, err := json.Marshal(&state)
		require.NoError(t, err)

		assert.Equal(t, b, []byte(`{"message_timout":10000000000,"woke_up_timeout":15000000000,"scale":10,"avatar":{"x":10,"y":10,"w":5,"h":5}}`))
	})
	t.Run("Unmarshal", func(t *testing.T) {
		b, err := json.Marshal(&state)
		require.NoError(t, err)

		var s State

		err = json.Unmarshal(b, &s)
		require.NoError(t, err)

		// Avatar has an incrementive internal id so we cannot compare directly
		// and this ID cannot be set so I have to compare the Bounds and then
		// assign the state.Avatar to the just Unmarshal s.Avatar
		if assert.Equal(t, s.Avatar.Bounds(), state.Avatar.Bounds()) {
			s.Avatar = state.Avatar
		}

		assert.Equal(t, s, state)
	})
}
