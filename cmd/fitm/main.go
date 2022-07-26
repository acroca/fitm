package main

import (
	"fmt"
	"log"
	"os"

	fitm "github.com/acroca/fitm/pkg"
	cli "github.com/urfave/cli/v2"
)

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
				Name:    "init",
				Aliases: []string{"i"},
				Usage:   "Initialises fitm to be ready to start.",
				Action:  fitm.InitAction,
			},
			{
				Name:   "up",
				Usage:  "runs the required components.",
				Flags:  []cli.Flag{},
				Action: fitm.UpAction,
			},
			{
				Name:   "down",
				Usage:  "stops all components.",
				Action: fitm.DownAction,
			},
			{
				Name:  "buckets",
				Usage: "bucket operations.",

				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "vault_address",
						Value:   "http://localhost:8200",
						Usage:   "Address where Vault is located",
						EnvVars: []string{"VAULT_ADDRESS"},
					},
					&cli.StringFlag{
						Name:     "vault_token",
						Usage:    "Vault token",
						EnvVars:  []string{"VAULT_TOKEN"},
						Required: true,
					},
				},

				Subcommands: []*cli.Command{
					{
						Name:   "list",
						Usage:  "list buckets.",
						Action: fitm.ListBucketsAction,
					},
					{
						Name:   "create",
						Usage:  "create a bucket.",
						Action: fitm.CreateBucketsAction,
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:     "id",
								Usage:    "Bucket ID",
								Required: true,
							},
						},
					},
					{
						Name:   "delete",
						Usage:  "delete a bucket.",
						Action: fitm.DeleteBucketsAction,
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:     "id",
								Usage:    "Bucket ID",
								Required: true,
							},
						},
					},
				},
			},
			{
				Name:  "users",
				Usage: "User operations.",

				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "vault_address",
						Value:   "http://localhost:8200",
						Usage:   "Address where Vault is located",
						EnvVars: []string{"VAULT_ADDRESS"},
					},
					&cli.StringFlag{
						Name:     "vault_token",
						Usage:    "Vault token",
						EnvVars:  []string{"VAULT_TOKEN"},
						Required: true,
					},
				},

				Subcommands: []*cli.Command{
					{
						Name:   "list",
						Usage:  "list users.",
						Action: fitm.ListUsersAction,
					},
					{
						Name:   "token",
						Usage:  "Generates a token to access specific buckets.",
						Action: fitm.TokenAction,
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:     "id",
								Usage:    "User ID",
								Required: true,
							},
							&cli.StringSliceFlag{
								Name:     "bucket-ids",
								Usage:    "Bucket IDs",
								Required: true,
							},
						},
					},
					{
						Name:   "create",
						Usage:  "create a user.",
						Action: fitm.CreateUsersAction,
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:     "id",
								Usage:    "User ID",
								Required: true,
							},
						},
					},
					{
						Name:   "delete",
						Usage:  "delete a user.",
						Action: fitm.DeleteUsersAction,
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:     "id",
								Usage:    "User ID",
								Required: true,
							},
						},
					},
				},
			},
			{
				Name:  "acl",
				Usage: "Access control operations.",

				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "vault_address",
						Value:   "http://localhost:8200",
						Usage:   "Address where Vault is located",
						EnvVars: []string{"VAULT_ADDRESS"},
					},
					&cli.StringFlag{
						Name:     "vault_token",
						Usage:    "Vault token",
						EnvVars:  []string{"VAULT_TOKEN"},
						Required: true,
					},
				},

				Subcommands: []*cli.Command{
					{
						Name:   "grant",
						Usage:  "Grants access to users in buckets.",
						Action: fitm.GrantAccessAction,
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:     "user-id",
								Usage:    "User ID",
								Required: true,
							},
							&cli.StringFlag{
								Name:     "bucket-id",
								Usage:    "Bucket ID",
								Required: true,
							},
						},
					},
					{
						Name:   "revoke",
						Usage:  "Revokes access of users in buckets.",
						Action: fitm.RevokeAccessAction,
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:     "user-id",
								Usage:    "User ID",
								Required: true,
							},
							&cli.StringFlag{
								Name:     "bucket-id",
								Usage:    "Bucket ID",
								Required: true,
							},
						},
					},
				},
			},
			{
				Name:  "browser",
				Usage: "Browser operations.",

				Subcommands: []*cli.Command{
					{
						Name:   "install",
						Usage:  "Installs embedded browser.",
						Action: fitm.BrowserInstallAction,
					},
					{
						Name:   "open",
						Usage:  "Opens an embedded browser configured and ready to use.",
						Action: fitm.BrowserRunAction,
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:     "bucket-id",
								Usage:    "Bucket ID",
								Required: true,
							},
							&cli.StringFlag{
								Name:     "token",
								Usage:    "Token",
								Required: true,
							},
						},
					},
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
