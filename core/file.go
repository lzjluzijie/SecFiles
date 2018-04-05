package core

import (
	"io"
	"os"

	"golang.org/x/crypto/sha3"
)

type File struct {
	Name string
	Path string
	Hash []byte
	Size int64
}

func OpenFile(path string) (f *File, err error) {
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

	f = &File{
		Name: file.Name(),
		Path: path,
		Hash: h2[:],
		Size: fs.Size(),
	}
	return
}
