package commands

import (
	"log"

	"github.com/lzjluzijie/secfiles/core"
	"github.com/lzjluzijie/secfiles/storage/baiduwangpan"
	"github.com/urfave/cli"
)

func Upload(ctx *cli.Context) (err error) {
	bduss := ctx.Args().Get(0)
	key := []byte(ctx.Args().Get(1))
	path := ctx.Args().Get(2)

	b, err := baiduwangpan.NewBaiduWangPan(bduss, baiduwangpan.App_id, key)
	if err != nil {
		log.Fatalln(err.Error())
		return
	}

	f, err := core.OpenSeed(path)
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
