package mcmd

import "github.com/urfave/cli/v2"

type IntegrateWithUrfave interface {
	AdaptToCLI(app *cli.App)
}
