package commands

import (
	"crypto/aes"
	"crypto/cipher"
	"io"
	"log"
	"os"

	"github.com/urfave/cli"
)

func Decrypt(ctx *cli.Context) (err error) {
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
