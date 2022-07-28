package fitm

import (
	"errors"
	"fmt"
	"runtime"

	"github.com/tj/go-update"
	"github.com/tj/go-update/stores/github"
	cli "github.com/urfave/cli/v2"
)

var version string
var updateManager *update.Manager

func SetVersion(v string) {
	version = v
	updateManager = &update.Manager{
		Command: "fitm",
		Store: &github.Store{
			Owner:   "acroca",
			Repo:    "fitm",
			Version: v,
		},
	}
}

func CheckVersionAction(ctx *cli.Context) error {
	// allow local development
	if version == "master" {
		return nil
	}

	_, err := getLatestValidVersion()
	return err
}

func UpdateAction(ctx *cli.Context) error {
	latest, err := getLatestValidVersion()
	if err != nil {
		return err
	}
	tar, err := latest.Download()
	if err != nil {
		return err
	}
	updateManager.Install(tar)
	fmt.Println("Updated!")
	return nil
}

func VersionAction(ctx *cli.Context) error {
	fmt.Printf(version)
	return nil
}

func getLatestValidVersion() (*update.Asset, error) {
	releases, err := updateManager.LatestReleases()
	if err != nil {
		return nil, err
	}
	if len(releases) == 0 {
		fmt.Println("No updates.")
		return nil, nil
	}
	for _, release := range releases {
		a := release.FindTarball(runtime.GOOS, runtime.GOARCH)
		if a != nil {
			return a, nil
		}
	}
	return nil, errors.New("no binary for your system")
}
