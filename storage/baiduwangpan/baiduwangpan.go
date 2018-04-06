package baiduwangpan

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"time"

	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"
	"strings"

	"github.com/bitly/go-simplejson"
	"github.com/lzjluzijie/secfiles/core"
)

var blocksize = int64(1024 * 1024)

var app_id = "260149"

var pcsURL = &url.URL{
	Scheme: "http",
	Host:   "pcs.baidu.com",
}

var panURL = &url.URL{
	Scheme: "http",
	Host:   "pan.baidu.com",
}

type BaiduWangPan struct {
	*http.Client

	key    []byte
	bduss  string
	app_id string
}

type baiduFile struct {
	*os.File
	path string
	name string
	size int64
}

type finish struct {
	id  int64
	err error
}

var pcsFileURL = "https://pcs.baidu.com/rest/2.0/pcs/file"

func (b *BaiduWangPan) Put(f *core.File) (err error) {
	file, err := os.Open(f.Path)
	if err != nil {
		return
	}

	//Encrypt file
	block, err := aes.NewCipher(b.key)
	if err != nil {
		return
	}

	iv := make([]byte, 16)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return err
	}

	stream := cipher.NewOFB(block, iv)
	reader := &cipher.StreamReader{
		S: stream,
		R: file,
	}

	//Ready to upload
	mr := core.NewMultipartReader()
	form := fmt.Sprintf("--%s\r\nContent-Disposition: form-data; name=\"%s\"; filename=\"%s\"\r\n\r\n", mr.Boundary, "file", fmt.Sprintf("%x.sfs", f.Hash))
	mr.AddReader(strings.NewReader(form), int64(len(form)))
	mr.AddReader(bytes.NewReader(iv), 16)
	mr.AddReader(reader, f.Size)

	req, err := http.NewRequest("POST", pcsFileURL, mr)
	if err != nil {
		return
	}
	mr.SetupHTTPRequest(req)
	v := req.URL.Query()
	v.Add("app_id", b.app_id)
	v.Add("method", "upload")
	v.Add("path", "/secfiles/"+fmt.Sprintf("%x.sfs", f.Hash))
	v.Add("ondup", "newcopy")
	req.URL.RawQuery = v.Encode()

	go func() {
		t := time.Now()
		for {
			time.Sleep(time.Second)
			readed := mr.Readed()
			if readed == f.Size {
				return
			}

			readed = readed / 1024

			log.Printf("%s uploaded:%dKB, speed:%dKBps", f.Name, readed, readed/int64(time.Since(t).Seconds()))
		}
	}()

	resp, err := b.Do(req)
	if err != nil {
		return
	}

	log.Println(resp.Header)

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	log.Println(string(data))

	return
}

func (b *BaiduWangPan) Get(f *core.File) (err error) {
	path := fmt.Sprintf("/secfiles/%x.sfs", f.Hash)
	name := fmt.Sprintf("%x.sfs", f.Hash)

	//Get size
	u := fmt.Sprintf("%s?app_id=%s&method=meta&path=%s", pcsFileURL, b.app_id, path)
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return
	}

	resp, err := b.Do(req)
	if err != nil {
		return
	}

	log.Println(resp)

	j, err := simplejson.NewFromReader(resp.Body)
	if err != nil {
		return
	}

	size := j.Get("list").GetIndex(0).Get("size").MustInt64()
	log.Printf("%s size: %d", name, size)

	file, err := os.Create(name)
	if err != nil {
		return
	}

	//Download file
	taskChan := make(chan int64)
	finishChan := make(chan finish)
	go func() {
		for id := int64(0); id*blocksize < size; id++ {
			taskChan <- id
		}
	}()

	for i := 0; i < 64; i++ {
		go func() {
			for id := range taskChan {
				start := id * blocksize
				end := start + blocksize - 1
				if end > size {
					end = size - 1
				}

				u = fmt.Sprintf("%s?app_id=%s&method=download&path=%s", pcsFileURL, b.app_id, path)

				req, err = http.NewRequest("GET", u, nil)
				if err != nil {
					finishChan <- finish{
						id:  id,
						err: err,
					}
					continue
				}

				req.Header.Add("Range", fmt.Sprintf("bytes=%d-%d", start, end))

				t := time.Now()
				resp, err = b.Do(req)
				if err != nil {
					finishChan <- finish{
						id:  id,
						err: err,
					}
					continue
				}

				data, err := ioutil.ReadAll(resp.Body)

				if err != nil {
					finishChan <- finish{
						id:  id,
						err: err,
					}
					continue
				}

				log.Printf("Block%d %dbytes %f", id, len(data), time.Since(t).Seconds())

				_, err = file.WriteAt(data, start)
				if err != nil {
					finishChan <- finish{
						id:  id,
						err: err,
					}
					continue
				}

				finishChan <- finish{
					id:  id,
					err: nil,
				}
			}
		}()
	}

	//Waiting download
	for finished := int64(0); finished*blocksize < size; {
		f := <-finishChan
		if f.err != nil {
			log.Fatalf("ID%d error: %s", f.id, err.Error())
			taskChan <- f.id
		}
		log.Printf("ID%d finished", f.id)
		finished++
	}

	//Download finished, start decrypt
	log.Printf("Download %s finished, begin to decrypt", name)
	b.Decrypt(name, f.Name)
	log.Println("Finish")
	return
}

func (b *BaiduWangPan) Decrypt(in, out string) (err error) {
	inFile, err := os.Open(in)
	if err != nil {
		log.Fatalln(err.Error())
		return
	}
	defer inFile.Close()

	block, err := aes.NewCipher(b.key)
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

func NewBaiduWangPan(bduss string, key []byte) (b *BaiduWangPan, err error) {
	cookie := &http.Cookie{
		Name:  "BDUSS",
		Value: bduss,
	}

	jar, err := cookiejar.New(nil)
	if err != nil {
		return
	}

	jar.SetCookies(pcsURL, []*http.Cookie{
		cookie,
	})

	jar.SetCookies(panURL, []*http.Cookie{
		cookie,
	})

	b = &BaiduWangPan{
		Client: &http.Client{
			Jar: jar,
		},
		app_id: app_id,
		bduss:  bduss,
		key:    key,
	}
	return
}
