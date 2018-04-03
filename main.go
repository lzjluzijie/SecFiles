package main

import (
	"os"

	"log"

	"github.com/lzjluzijie/secfiles/storage/baiduwangpan"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "SecFiles"
	app.Usage = "Save your files!"
	app.Author = "Halulu"
	app.Version = "0.0.1"

	app.Commands = []cli.Command{
		{
			Name:    "download",
			Aliases: []string{"dl"},
			Usage:   "Just download",
			Action:  Test,
		},
	}

	app.Run(os.Args)
}

func Test(ctx *cli.Context) (err error) {
	bduss := ctx.Args().Get(0)
	path := ctx.Args().Get(1)
	name := ctx.Args().Get(2)

	b, err := baiduwangpan.NewBaiduWangPan(bduss)
	if err != nil {
		log.Fatalln(err.Error())
		return
	}
	err = b.Get(path, name)
	if err != nil {
		log.Fatalln(err.Error())
		return
	}
	return
}
