package main

import (
	"os"

	"log"

	"crypto/aes"
	"crypto/cipher"
	"io"

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
		{
			Name:    "decrypt",
			Aliases: []string{"dec"},
			Usage:   "Just decrypt",
			Action:  decrypt,
		},
	}

	app.Run(os.Args)
}

func download(ctx *cli.Context) (err error) {
	bduss := ctx.Args().Get(0)
	key := []byte(ctx.Args().Get(1))
	hexHash := ctx.Args().Get(2)
	name := ctx.Args().Get(3)

	f, err := core.ParseFile(hexHash, name)
	if err != nil {
		log.Fatalln(err.Error())
		return
	}

	b, err := baiduwangpan.NewBaiduWangPan(bduss, key)
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

	b, err := baiduwangpan.NewBaiduWangPan(bduss, key)
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
