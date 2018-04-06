package main

import (
	"crypto/aes"
	"crypto/cipher"
	"io"
	"log"
	"os"

	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/lzjluzijie/base36"
	"github.com/lzjluzijie/secfiles/core"
	"github.com/lzjluzijie/secfiles/storage/baiduwangpan"
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
			Action:  add,
		},
		{
			Name:    "create",
			Aliases: []string{"create"},
			Usage:   "Just create",
			Action:  create,
		},
		{
			Name:    "put",
			Aliases: []string{"put"},
			Usage:   "Just put",
			Action:  put,
		},
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
		{
			Name:    "decrypt",
			Aliases: []string{"dec"},
			Usage:   "Just decrypt",
			Action:  decrypt,
		},
	}

	app.Run(os.Args)
}

var app_id = "260149"

func add(ctx *cli.Context) (err error) {
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

	wd, err := os.Getwd()
	if err != nil {
		log.Fatalln(err.Error())
		return
	}

	for _, p := range ctx.Args() {
		fi, err := os.Stat(p)
		if err != nil {
			log.Fatalln(err.Error())
			continue
		}

		if !fi.IsDir() {
			s, err := core.OpenSeed(p)
			if err != nil {
				log.Fatalln(err.Error())
				continue
			}

			seedS.Seeds = append(seedS.Seeds, s)
			continue
		}

		seeds, err := core.GetSeeds(wd + string(os.PathSeparator) + fi.Name())
		if err != nil {
			log.Fatalln(err.Error())
			continue
		}

		seedS.Seeds = append(seedS.Seeds, seeds...)
	}

	data, err = json.MarshalIndent(seedS, "", "    ")
	if err != nil {
		log.Fatalln(err.Error())
		return
	}

	ioutil.WriteFile("seeds.json", data, 0600)
	return
}

func create(ctx *cli.Context) (err error) {
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

func put(ctx *cli.Context) (err error) {
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

	b, err := baiduwangpan.NewBaiduWangPan(seedS.BDUSS, app_id, key)
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

func download(ctx *cli.Context) (err error) {
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

	b, err := baiduwangpan.NewBaiduWangPan(bduss, app_id, key)
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

func upload(ctx *cli.Context) (err error) {
	bduss := ctx.Args().Get(0)
	key := []byte(ctx.Args().Get(1))
	path := ctx.Args().Get(2)

	b, err := baiduwangpan.NewBaiduWangPan(bduss, app_id, key)
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

func decrypt(ctx *cli.Context) (err error) {
	key := []byte(ctx.Args().Get(0))
	in := ctx.Args().Get(1)
	out := ctx.Args().Get(2)

	inFile, err := os.Open(in)
	if err != nil {
		log.Fatalln(err.Error())
		return
	}
	defer inFile.Close()

	block, err := aes.NewCipher(key)
	if err != nil {
		log.Fatalln(err.Error())
		return
	}

	iv := make([]byte, 16)
	_, err = inFile.Read(iv)
	if err != nil {
		log.Fatalln(err.Error())
		return
	}

	stream := cipher.NewOFB(block, iv)

	outFile, err := os.OpenFile(out, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalln(err.Error())
		return
	}

	defer outFile.Close()

	reader := &cipher.StreamReader{
		S: stream,
		R: inFile,
	}

	if _, err := io.Copy(outFile, reader); err != nil {
		log.Fatalln(err.Error())
		return err
	}

	return
}
