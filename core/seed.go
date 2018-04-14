package core

import (
	"io"
	"os"

	"errors"
	"fmt"

	"github.com/lzjluzijie/base36"
	"golang.org/x/crypto/sha3"
)

type Seed struct {
	Name    string
	Path    string
	Size    int64
	Hash    []byte `json:"-"`
	B36Hash string `json:"hash"`
}

func OpenSeed(path string) (s *Seed, err error) {
	file, err := os.Open(path)
	if err != nil {
		return
	}

	fs, err := file.Stat()
	if err != nil {
		return
	}

	h := sha3.New512()
	_, err = io.Copy(h, file)
	if err != nil {
		return
	}

	h1 := h.Sum(nil)
	h.Reset()
	h.Write(h1[:])
	h2 := h.Sum(nil)

	b36hash := base36.Encode(h2)

	s = &Seed{
		Name:    fs.Name(),
		Path:    path,
		Size:    fs.Size(),
		Hash:    h2[:],
		B36Hash: b36hash,
	}
	return
}

func ParseSeed(b36hash, name string) (f *Seed, err error) {
	hash := base36.Decode(b36hash)
	if len(hash) != 64 {
		err = errors.New(fmt.Sprintf("unknown hash length: %d", len(hash)))
		return
	}

	f = &Seed{
		Name:    name,
		Hash:    hash,
		B36Hash: b36hash,
	}
	return
}
