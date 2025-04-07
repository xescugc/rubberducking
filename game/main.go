package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/adrg/xdg"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/spf13/afero"
	"github.com/xescugc/go-flux/v2"
	"github.com/xescugc/rubberducking/src"
)

var (
	stateFile = path.Join(xdg.DataHome, src.AppName, "state.json")
)

func initFs(fs afero.Fs) error {
	err := fs.MkdirAll(filepath.Dir(stateFile), 0700)
	if err != nil {
		return fmt.Errorf("failed to MkdirAll: %w", err)
	}
	return nil
}

func main() {
	fs := afero.NewOsFs()
	initLog()
	initFs(fs)

	Logger.Info("Starting Game")

	port := os.Getenv("PORT")
	verbose := os.Getenv("VERBOSE")
	managerURL := os.Getenv("MANAGER_URL")

	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeDisabled)
	ebiten.SetWindowFloating(true)
	ebiten.SetWindowDecorated(false)
	ebiten.SetWindowSize(ebiten.Monitor().Size())

	d := flux.NewDispatcher[*Action]()

	store := NewStore(d, fs, time.Second*10, time.Second*15)
	defer func() {
		state := store.GetState()
		b, err := json.Marshal(&state)
		if err != nil {
			Logger.Error("Failed to Marshal State", "err", err)
			return
		}
		err = afero.WriteFile(fs, stateFile, b, 0644)
		if err != nil {
			Logger.Error("Failed to WriteFile State", "err", err)
			return
		}
	}()

	ad := NewActionDispatcher(d)
	g := NewGame(d, store, ad)

	go g.startHttpServer(port, verbose == "true")

	_, err := http.Post(managerURL+"/game/start", "application/json", nil)
	if err != nil {
		Logger.Error("Error on communicating to the manger", "err", err)
		return
	}
	if err := ebiten.RunGameWithOptions(g, &ebiten.RunGameOptions{
		ScreenTransparent: true,
		InitUnfocused:     true,
	}); err != nil {
		Logger.Error("Error on RunGameOptions", "err", err)
		return
	}
	Logger.Info("Exiting the game")
}
