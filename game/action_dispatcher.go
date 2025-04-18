package main

import "github.com/xescugc/go-flux/v2"

type ActionDispatcher struct {
	dispatcher *flux.Dispatcher[*Action]
}

func NewActionDispatcher(d *flux.Dispatcher[*Action]) *ActionDispatcher {
	return &ActionDispatcher{
		dispatcher: d,
	}
}

func (ac *ActionDispatcher) Dispatch(a *Action) {
	ac.dispatcher.Dispatch(a)
}

func (ac *ActionDispatcher) DragAvatar(x, y int) {
	da := NewDragAvatar(x, y)
	ac.Dispatch(da)
}

func (ac *ActionDispatcher) AddMessage(m string) {
	ama := NewAddMessage(m)
	ac.Dispatch(ama)
}

func (ac *ActionDispatcher) TPS() {
	tpsa := NewTPS()
	ac.Dispatch(tpsa)
}

func (ac *ActionDispatcher) Toggle() {
	ta := NewToggle()
	ac.Dispatch(ta)
}

func (ac *ActionDispatcher) MenuOpen(o bool) {
	moa := NewMenuOpen(o)
	ac.Dispatch(moa)
}

func (ac *ActionDispatcher) SetFocusMode(f bool) {
	fma := NewSetFocusMode(f)
	ac.Dispatch(fma)
}
