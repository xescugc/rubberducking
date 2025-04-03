package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/hajimehoshi/ebiten/v2"
	cli "github.com/urfave/cli/v3"
	"github.com/xescugc/go-flux/v2"
	"github.com/xescugc/rubberducking/game"
)

const (
	defaultPort = "6302"
)

var (
	Cmd = &cli.Command{
		Name:  "rubberducking",
		Usage: "Summon your personal Rubber Duck!",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "port", Value: defaultPort, Usage: "Open HTTP port to communicate with the duck!"},
			&cli.BoolFlag{Name: "verbose", Value: false, Usage: "Activate verbose mode to display logs and info"},
		},
		Commands: []*cli.Command{
			{
				Name:      "send-message",
				Usage:     "Sends a message to your Duck!",
				ArgsUsage: "[message]",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "message", Value: "", Usage: "Message to send to the duck! Can also be send as a new arg"},
				},
				Action: func(_ context.Context, cmd *cli.Command) error {
					cmr := game.CreateMessageRequest{Message: cmd.Args().Get(0)}
					b, err := json.Marshal(&cmr)
					if err != nil {
						return err
					}

					resp, err := http.NewRequest(http.MethodPost, "http://localhost:"+cmd.String("port")+"/messages", bytes.NewBuffer(b))
					if err != nil {
						return err
					}

					response, err := http.DefaultClient.Do(resp)
					if err != nil {
						return err
					}
					defer response.Body.Close()

					if response.StatusCode == http.StatusBadRequest {
						var eb game.ErrorResponse
						err := json.NewDecoder(response.Body).Decode(&eb)
						if err != nil {
							return err
						}
						return errors.New(eb.Error)
					}
					return nil
				},
			},
		},
		Action: func(_ context.Context, cmd *cli.Command) error {
			ebiten.SetWindowResizingMode(ebiten.WindowResizingModeDisabled)
			ebiten.SetWindowFloating(true)
			ebiten.SetWindowDecorated(false)
			ebiten.SetWindowSize(ebiten.Monitor().Size())

			d := flux.NewDispatcher[*game.Action]()

			g := game.NewGame(d, cmd.String("port"), cmd.Bool("verbose"))
			if err := ebiten.RunGameWithOptions(g, &ebiten.RunGameOptions{
				ScreenTransparent: true,
				InitUnfocused:     true,
			}); err != nil {
				return err
			}
			return nil
		},
	}
)
