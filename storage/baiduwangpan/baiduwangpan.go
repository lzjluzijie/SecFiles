package baiduwangpan

import (
	"net/http"
	"net/http/cookiejar"
	"net/url"

	"os"

	"fmt"
	"io/ioutil"
	"log"

	"github.com/bitly/go-simplejson"
	"time"
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
	bduss  string
	app_id string
	client Client
}

type Client struct {
	*http.Client
	app_id string
}

func (client *Client) NewRequest(method string, url string, q map[string]string) (req *http.Request, err error) {
	req, err = http.NewRequest(method, url, nil)
	if err != nil {
		return
	}

	v := req.URL.Query()
	v.Add("app_id", client.app_id)
	for key, value := range q {
		v.Add(key, value)
	}
	req.URL.RawQuery = v.Encode()

	return
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

func (b *BaiduWangPan) Get(path, name string) (err error) {
	//Get size
	req, err := b.client.NewRequest("GET", pcsFileURL, map[string]string{
		"method": "meta",
		"path":   path,
	})
	if err != nil {
		return
	}

	resp, err := b.client.Do(req)
	if err != nil {
		return
	}

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

	for i := 0; i < 4; i++ {
		go func() {
			for id := range taskChan {
				start := id * blocksize
				end := start + blocksize - 1
				if end > size {
					end = size -1
				}

				req, err = b.client.NewRequest("GET", pcsFileURL, map[string]string{
					"method": "download",
					"path":   path,
				})
				if err != nil {
					finishChan <- finish{
						id:  id,
						err: err,
					}
					continue
				}

				req.Header.Add("Range", fmt.Sprintf("bytes=%d-%d", start, end))

				t := time.Now()
				resp, err = b.client.Do(req)
				if err != nil {
					finishChan <- finish{
						id:  id,
						err: err,
					}
					continue
				}

				data, err := ioutil.ReadAll(resp.Body)
				if len(data)==65{
					log.Println(string(data))
				}

				if err != nil {
					finishChan <- finish{
						id:  id,
						err: err,
					}
					continue
				}

				log.Printf("Block%d %dbytes %f",id,len(data), time.Since(t).Seconds())

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

	c := Client{
		Client: &http.Client{
			Jar: jar,
		},
		app_id: app_id,
	}

	b = &BaiduWangPan{
		app_id: app_id,
		bduss:  bduss,
		client: c,
	}
	return
}
