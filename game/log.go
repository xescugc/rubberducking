package main

import (
	"os"
	"path"

	"log/slog"

	"github.com/adrg/xdg"
	"github.com/xescugc/rubberducking/src"
)

var (
	Logger *slog.Logger

	Level = new(slog.LevelVar)

	logFilePath = path.Join(xdg.StateHome, src.AppName, "game.log")

	logWriter = os.Stdout
)

func initLog() {
	if Logger == nil {
		Level.Set(slog.LevelInfo)

		Logger = slog.New(slog.NewTextHandler(logWriter, &slog.HandlerOptions{
			Level: Level,
		}))
	}
}
