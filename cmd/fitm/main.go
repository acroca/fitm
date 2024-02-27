package main

import (
	"fmt"
	"log"
	"os"

	fitm "github.com/acroca/fitm/internal/cli"
	cli "github.com/urfave/cli/v2"
)

var version = "master"

func assertErrorToNilf(message string, err error) {
	if err != nil {
		log.Fatalf(message, err)
	}
}

func todoAction(c *cli.Context) error {
	fmt.Println("TODO!")
	return nil
}

func main() {
	app := &cli.App{
		Name:  "fitm",
		Usage: "client for the fitm API",

		Commands: []*cli.Command{
			{
				Name:   "run",
				Usage:  "Runs the fitm service.",
				Action: fitm.RunAction,
			},
			{
				Name:  "version",
				Usage: "prints out tool version.",
				Action: func(ctx *cli.Context) error {
					fmt.Printf(version)
					return nil
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
