package cmd

import (
	"context"

	"github.com/spf13/afero"
	"github.com/xescugc/rubberducking/src"

	cli "github.com/urfave/cli/v3"
)

const (
	defaultPort = "6302"
)

var (
	Cmd = &cli.Command{
		Name:  src.AppName,
		Usage: "Summon your personal Rubber Duck!",
		Flags: []cli.Flag{
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
				Action: func(ctx context.Context, cmd *cli.Command) error {
					msg := cmd.Args().Get(0)
					if cmd.String("message") != "" {
						msg = cmd.String("message")
					}

					return src.SendMessage(ctx, afero.NewOsFs(), msg)
				},
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return src.Manager(ctx, afero.NewOsFs(), cmd.Bool("verbose"))
		},
	}
)
