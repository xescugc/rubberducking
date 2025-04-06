package main

import (
	"net/http"
	"os"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/xescugc/go-flux/v2"
)

func main() {
	initLog()

	Logger.Info("Starting Game")

	port := os.Getenv("PORT")
	verbose := os.Getenv("VERBOSE")
	managerURL := os.Getenv("MANAGER_URL")

	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeDisabled)
	ebiten.SetWindowFloating(true)
	ebiten.SetWindowDecorated(false)
	ebiten.SetWindowSize(ebiten.Monitor().Size())

	d := flux.NewDispatcher[*Action]()

	store := NewStore(d, time.Second*10, time.Second*15)
	ad := NewActionDispatcher(d)
	g := NewGame(d, store, ad)

	go g.startHttpServer(port, verbose == "true")

	_, err := http.Post(managerURL+"/game/start", "application/json", nil)
	if err != nil {
		Logger.Error("Error on communicating to the manger", "err", err)
		os.Exit(1)
	}
	if err := ebiten.RunGameWithOptions(g, &ebiten.RunGameOptions{
		ScreenTransparent: true,
		InitUnfocused:     true,
	}); err != nil {
		Logger.Error("Error on RunGameOptions", "err", err)
		os.Exit(1)
	}
	Logger.Info("Exiting the game")
}
