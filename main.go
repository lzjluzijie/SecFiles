package main

import (
	"os"

	"log"

	"github.com/lzjluzijie/secfiles/core"
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
			Aliases: []string{"d"},
			Usage:   "Just download",
			Action:  download,
		},
		{
			Name:    "upload",
			Aliases: []string{"u"},
			Usage:   "Just upload",
			Action:  upload,
		},
	}

	app.Run(os.Args)
}

func upload(ctx *cli.Context) (err error) {
	bduss := ctx.Args().Get(0)
	path := ctx.Args().Get(1)

	b, err := baiduwangpan.NewBaiduWangPan(bduss)
	if err != nil {
		log.Fatalln(err.Error())
		return
	}

	f, err := core.OpenFile(path)
	if err != nil {
		log.Fatalln(err.Error())
		return
	}

	err = b.Put(f)
	if err != nil {
		log.Fatalln(err.Error())
		return
	}
	return
}

func download(ctx *cli.Context) (err error) {
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
