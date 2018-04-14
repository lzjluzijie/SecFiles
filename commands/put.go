package commands

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/lzjluzijie/base36"
	"github.com/lzjluzijie/secfiles/core"
	"github.com/lzjluzijie/secfiles/storage/baiduwangpan"
	"github.com/urfave/cli"
)

func Put(ctx *cli.Context) (err error) {
	seedS := new(core.SeedS)

	data, err := ioutil.ReadFile("seeds.json")
	if err != nil {
		log.Fatalln(err.Error())
		return
	}

	err = json.Unmarshal(data, seedS)
	if err != nil {
		log.Fatalln(err.Error())
		return
	}

	key := base36.Decode(seedS.B36Key)
	if len(key) != 32 {
		err = errors.New(fmt.Sprintf("unknown key length: %d", len(key)))
	}

	b, err := baiduwangpan.NewBaiduWangPan(seedS.BDUSS, baiduwangpan.App_id, key)
	if err != nil {
		log.Fatalln(err.Error())
		return
	}

	for _, seed := range seedS.Seeds {
		err = b.Put(seed)
		if err != nil {
			log.Fatalln(err.Error())
			continue
		}
	}

	return
}
