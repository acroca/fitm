package fitm

import (
	_ "embed"
	"os"
	"os/exec"
	"path"
	"text/template"

	cli "github.com/urfave/cli/v2"
)

var (
	//go:embed docker-compose.yml.tmpl
	dockerComposeTemplate string

	configDir             = path.Join(os.Getenv("HOME"), ".config", "fitm")
	dockerComposeFileName = path.Join(configDir, "docker-compose.yml")
)

func InitAction(c *cli.Context) error {
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	dockerComposeFile, err := os.Create(dockerComposeFileName)
	if err != nil {
		return err
	}

	tpl, err := template.New("docker_compose.yml").Parse(dockerComposeTemplate)
	if err != nil {
		return err
	}
	if err := tpl.Execute(dockerComposeFile, nil); err != nil {
		return err
	}

	return nil
}

func UpAction(c *cli.Context) error {
	cmd := exec.Command("docker", "compose", "-f", dockerComposeFileName, "up", "-d")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func DownAction(c *cli.Context) error {
	cmd := exec.Command("docker", "compose", "-f", dockerComposeFileName, "down")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
