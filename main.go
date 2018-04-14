package main

import (
	"os"

	"github.com/lzjluzijie/secfiles/commands"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "SecFiles"
	app.Usage = "Save your files!"
	app.Author = "Halulu"
	app.Version = "0.1.0"

	app.Commands = []cli.Command{
		{
			Name:    "add",
			Aliases: []string{"add"},
			Usage:   "Just add",
			Action:  commands.Add,
		},
		{
			Name:    "create",
			Aliases: []string{"create"},
			Usage:   "Just create",
			Action:  commands.Create,
		},
		{
			Name:    "put",
			Aliases: []string{"put"},
			Usage:   "Just put",
			Action:  commands.Put,
		},
		{
			Name:    "download",
			Aliases: []string{"d"},
			Usage:   "Just download",
			Action:  commands.Download,
		},
		{
			Name:    "upload",
			Aliases: []string{"u"},
			Usage:   "Just upload",
			Action:  commands.Upload,
		},
		{
			Name:    "decrypt",
			Aliases: []string{"dec"},
			Usage:   "Just decrypt",
			Action:  commands.Decrypt,
		},
	}

	app.Run(os.Args)
}
