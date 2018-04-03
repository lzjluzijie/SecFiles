package baiduwangpan

import (
	"net/http"
	"net/http/cookiejar"
	"net/url"

	"os"

	"github.com/bitly/go-simplejson"
	"log"
	"io"
)

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
	log.Println(v)

	return
}

type baiduFile struct {
	*os.File
	path string
	name string
	size int64
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

	//Download file
	req, err = b.client.NewRequest("GET", pcsFileURL, map[string]string{
		"method": "download",
		"path":   path,
	})
	if err != nil {
		return
	}

	resp, err = b.client.Do(req)
	if err != nil {
		return
	}

	file, err := os.Create(name)
	if err != nil {
		return
	}

	_, err = io.Copy(file, resp.Body)

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
