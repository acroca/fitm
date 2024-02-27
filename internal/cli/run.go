package fitm

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"

	cli "github.com/urfave/cli/v2"
)

func RunAction(c *cli.Context) error {
	if err := checkDependencies(); err != nil {
		return fmt.Errorf("Error checking dependencies: %w", err)
	}

	curDir, err := os.Getwd()
	if err != nil {
		return err
	}
	confdir, err := getConfdir()
	if err != nil {
		return err
	}

	cmd := exec.Command(
		"mitmdump",

		"-s", path.Join(curDir, "fitm.py"),

		// Sets the confdir to a temp directory to not polute the user's home
		"--set", fmt.Sprintf("confdir=%v", confdir),
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		return err
	}

	return nil
}

func checkDependencies() error {
	err := exec.Command("which", "mitmdump").Run()
	if err != nil {
		return errors.New("mitmdump not found.")
	}
	return nil
}

func getConfdir() (string, error) {
	confdir := os.Getenv("MITMPROXY_CONFDIR")
	if confdir == "" {
		tmp, err := os.MkdirTemp("", "")
		if err != nil {
			return "", err
		}
		confdir = tmp
	}
	return confdir, nil
}
