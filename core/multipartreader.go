package core

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
)

/*
魔改multipartreader
https://github.com/iikira/BaiduPCS-Go/blob/master/requester/multipartreader/multipartreader.go
iikira
Apache License 2.0
*/

type MultipartReader struct {
	contentType string
	boundary    string

	readers     []io.Reader
	length      int64
	multiReader io.Reader
	readed      int64
}

func NewMultipartReader() (mr *MultipartReader) {
	buf := bytes.NewBuffer(nil)
	writer := multipart.NewWriter(buf)

	formBody := buf.String()
	formClose := "\r\n--" + writer.Boundary() + "--\r\n"

	bodyReader := strings.NewReader(formBody)
	closeReader := strings.NewReader(formClose)

	mr = &MultipartReader{
		contentType: writer.FormDataContentType(),
		boundary:    writer.Boundary(),
		readers:     []io.Reader{bodyReader, closeReader},
		length:      int64(len(formBody) + len(formClose)),
	}
	return
}

func (mr *MultipartReader) AddReader(r io.Reader, length int64) {
	i := len(mr.readers)
	mr.readers = append(mr.readers[:i-1], r, mr.readers[i-1])
	mr.length = mr.length + length
}

func (mr *MultipartReader) AddFile(file *os.File) (err error) {
	fs, err := file.Stat()
	if err != nil {
		return
	}

	form := fmt.Sprintf("--%s\r\nContent-Disposition: form-data; name=\"%s\"; filename=\"%s\"\r\n\r\n", mr.boundary, "file", fs.Name())
	mr.AddReader(strings.NewReader(form), int64(len(form)))
	mr.AddReader(file, fs.Size())
	return
}

func (mr *MultipartReader) SetupHTTPRequest(req *http.Request) {
	mr.multiReader = io.MultiReader(mr.readers...)

	req.Header.Add("Content-Type", mr.contentType)
	req.ContentLength = mr.length
}

func (mr *MultipartReader) Read(p []byte) (n int, err error) {
	n, err = mr.multiReader.Read(p)
	atomic.AddInt64(&mr.readed, int64(n))
	return n, err
}

func (mr *MultipartReader) Readed() int64 {
	return atomic.LoadInt64(&mr.readed)
}
