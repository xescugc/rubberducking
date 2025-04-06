package main

import (
	"os"
	"path"

	"log/slog"

	"github.com/adrg/xdg"
)

var (
	AppName = "rubberducking"

	Logger *slog.Logger

	Level = new(slog.LevelVar)

	logFilePath = path.Join(xdg.StateHome, AppName, "game.log")

	logWriter = os.Stdout
)

func initLog() {
	Level.Set(slog.LevelInfo)

	Logger = slog.New(slog.NewTextHandler(logWriter, &slog.HandlerOptions{
		Level: Level,
	}))
}
