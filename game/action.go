package main

type Type int

//go:generate enumer -type=Type -transform=snake -output=action_type_string.go -json

const (
	DragAvatar Type = iota
	AddMessage
	TPS
	Toggle
)

type Action struct {
	Type Type

	DragAvatar DragAvatarPayload
	AddMessage AddMessagePayload
	TPS        TPSPayload
	Toggle     TogglePayload
}

type DragAvatarPayload struct {
	X int
	Y int
}

func NewDragAvatar(x, y int) *Action {
	return &Action{
		Type: DragAvatar,
		DragAvatar: DragAvatarPayload{
			X: x,
			Y: y,
		},
	}
}

type AddMessagePayload struct {
	Message string
}

func NewAddMessage(m string) *Action {
	return &Action{
		Type: AddMessage,
		AddMessage: AddMessagePayload{
			Message: m,
		},
	}
}

type TPSPayload struct{}

func NewTPS() *Action {
	return &Action{
		Type: TPS,
	}
}

type TogglePayload struct{}

func NewToggle() *Action {
	return &Action{
		Type: Toggle,
	}
}
