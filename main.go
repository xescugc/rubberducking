package main

import (
	"context"
	"log"
	"os"

	"github.com/xescugc/rubberducking/cmd"
)

func main() {
	err := cmd.Cmd.Run(context.TODO(), os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
