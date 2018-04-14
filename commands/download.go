package commands

import (
	"errors"
	"fmt"
	"log"

	"github.com/lzjluzijie/base36"
	"github.com/lzjluzijie/secfiles/core"
	"github.com/lzjluzijie/secfiles/storage/baiduwangpan"
	"github.com/urfave/cli"
)

func Download(ctx *cli.Context) (err error) {
	bduss := ctx.Args().Get(0)
	b36key := ctx.Args().Get(1)
	hexHash := ctx.Args().Get(2)
	name := ctx.Args().Get(3)

	key := base36.Decode(b36key)
	if len(key) != 32 {
		err = errors.New(fmt.Sprintf("unknown key length: %d", len(key)))
	}

	f, err := core.ParseSeed(hexHash, name)
	if err != nil {
		log.Fatalln(err.Error())
		return
	}

	b, err := baiduwangpan.NewBaiduWangPan(bduss, baiduwangpan.App_id, key)
	if err != nil {
		log.Fatalln(err.Error())
		return
	}

	err = b.Get(f)
	if err != nil {
		log.Fatalln(err.Error())
		return
	}

	return
}
