package baiduwangpan

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"time"

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
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", f.Path)
	if err != nil {
		return
	}

	file, err := os.Open(f.Path)
	if err != nil {
		return
	}

	_, err = io.Copy(part, file)

	err = writer.Close()
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", pcsFileURL, body)
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	v := req.URL.Query()
	v.Add("app_id", b.app_id)
	v.Add("method", "upload")
	v.Add("path", "/"+f.Name)
	v.Add("ondup", "newcopy")
	req.URL.RawQuery = v.Encode()

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

func (b *BaiduWangPan) Get(path, name string) (err error) {
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
				if len(data) == 65 {
					log.Println(string(data))
				}

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

	for finished := int64(0); finished*blocksize < size; {
		f := <-finishChan
		if f.err != nil {
			log.Fatalf("ID%d error: %s", f.id, err.Error())
			taskChan <- f.id
		}
		log.Printf("ID%d finished", f.id)
		finished++
	}

	return
}

func NewBaiduWangPan(bduss string) (b *BaiduWangPan, err error) {
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
	}
	return
}
