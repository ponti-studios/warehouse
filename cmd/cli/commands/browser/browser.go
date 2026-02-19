package browser

import (
	"fmt"
	"os"

	"github.com/urfave/cli/v2"
)

func Run() error {
	app := &cli.App{
		Name:  "browser",
		Usage: "Browser automation tools using Chrome DevTools Protocol",
		Commands: []*cli.Command{
			{
				Name:  "start",
				Usage: "Start Chrome with remote debugging on :9222",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "profile",
						Usage: "Copy your default Chrome profile (cookies, logins)",
					},
				},
				Action: func(c *cli.Context) error {
					if c.Bool("profile") {
						return runStartProfile()
					}
					return runStart()
				},
			},
			{
				Name:      "navigate",
				Usage:     "Navigate to a URL",
				ArgsUsage: "<url>",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "new",
						Usage: "Open in new tab",
					},
				},
				Action: func(c *cli.Context) error {
					url := c.Args().First()
					if url == "" {
						return fmt.Errorf("URL is required")
					}
					return runNavigate(url, c.Bool("new"))
				},
			},
			{
				Name:      "eval",
				Usage:     "Execute JavaScript in the active page context",
				ArgsUsage: "<code>",
				Action: func(c *cli.Context) error {
					code := c.Args().First()
					if code == "" {
						return fmt.Errorf("JavaScript code is required")
					}
					return runEval(code)
				},
			},
			{
				Name:  "screenshot",
				Usage: "Take a screenshot of the current viewport",
				Action: func(c *cli.Context) error {
					return runScreenshot()
				},
			},
			{
				Name:      "pick",
				Usage:     "Interactive element picker - click to select elements",
				ArgsUsage: "<message>",
				Action: func(c *cli.Context) error {
					message := c.Args().First()
					if message == "" {
						message = "Click an element"
					}
					return runPick(message)
				},
			},
			{
				Name:  "cookies",
				Usage: "Get cookies from the current page",
				Action: func(c *cli.Context) error {
					return runCookies()
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		return err
	}
	return nil
}
