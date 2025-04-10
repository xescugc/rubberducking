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

		assert.Equal(t, b, []byte(`{"message_timeout":10000000000,"message_max_line_characters":0,"message_max_lines":0,"woke_up_timeout":15000000000,"scale":10,"avatar":{"x":10,"y":10,"w":5,"h":5}}`))
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

func Test_truncateMessage(t *testing.T) {
	// On this test case we assume the ML and MLC are equal to 5
	tests := map[string]struct {
		in  string
		out string
	}{
		"Short": {
			in:  "msg",
			out: "msg",
		},
		"Short+NewLine": {
			in:  "msg\nmsg2",
			out: "msg\nmsg2",
		},
		"NewLines": {
			in:  "msg msg2",
			out: "msg\nmsg2",
		},
		"TooLongContinuous": {
			in:  "msgmsg2",
			out: "msgm-\nsg2",
		},
		"TooTooLongContinuous": {
			in:  "msgmsg2msg3msg4msg5",
			out: "msgm-\nsg2m-\nsg3m-\nsg4m-\nsg5",
		},
		"TooLongWithSpaces": {
			in:  "msg msg2msg3msg4msg5",
			out: "msg\nmsg2-\nmsg3-\nmsg4-\nmsg5",
		},
		"TooLongWithMultipleSpaces": {
			in:  "msg msg2msg3 msg4msg5",
			out: "msg\nmsg2-\nmsg3\nmsg4-\nmsg5",
		},
		"TooManyLines": {
			in:  "msgmsg2msg3msg4msg5msg6",
			out: "msgm-\nsg2m-\nsg3m-\nsg4m-\nsg5m-\n...",
		},
	}

	for tn, tt := range tests {
		t.Run(tn, func(t *testing.T) {
			out := truncateMessage(tt.in, 5, 5)
			assert.Equal(t, tt.out, out)
		})
	}
}
