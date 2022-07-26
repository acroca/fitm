package fitm

import (
	_ "embed"
	"os"
	"os/exec"
	"path"
	"text/template"

	"github.com/playwright-community/playwright-go"
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

func BrowserInstallAction(c *cli.Context) error {
	return playwright.Install(&playwright.RunOptions{
		Browsers: []string{"chromium"},
	})
}

func BrowserRunAction(c *cli.Context) error {
	bucket := c.String("bucket-id")
	token := c.String("token")

	pw, err := playwright.Run()
	if err != nil {
		return err
	}
	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(false),
		Proxy: &playwright.BrowserTypeLaunchOptionsProxy{
			Server:   playwright.String("localhost:8080"),
			Username: playwright.String(bucket),
			Password: playwright.String(token),
		},
	})
	if err != nil {
		return err
	}
	_, err = browser.NewPage()
	if err != nil {
		return err
	}
	onCloseWasCalled := make(chan bool, 1)
	onClose := func() {
		onCloseWasCalled <- true
	}
	browser.On("close", onClose)
	<-onCloseWasCalled

	return nil
}
