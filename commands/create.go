package commands

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"time"

	"github.com/lzjluzijie/base36"
	"github.com/lzjluzijie/secfiles/core"
	"github.com/urfave/cli"
)

func Create(ctx *cli.Context) (err error) {
	bduss := ctx.Args().Get(0)
	b36key := ctx.Args().Get(1)

	key := base36.Decode(b36key)

	if len(key) != 32 {
		err = errors.New(fmt.Sprintf("unknown key length: %d", len(key)))

		//gen key
		key = make([]byte, 32)
		if _, err := io.ReadFull(rand.Reader, key); err != nil {
			return err
		}

		b36key = base36.Encode(key)
	}

	seeds := &core.SeedS{
		Name:      "test",
		BDUSS:     bduss,
		Key:       key,
		B36Key:    b36key,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	data, err := json.MarshalIndent(seeds, "", "    ")
	if err != nil {
		log.Fatalln(err.Error())
		return
	}

	ioutil.WriteFile("seeds.json", data, 0600)
	return
}
