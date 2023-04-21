package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

var envCookie = os.Getenv("cookie")
var envAI = os.Getenv("ai")

var client http.Client

type transport struct {
	T http.RoundTripper
}

func (t *transport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Add("cookie", envCookie)
	return t.T.RoundTrip(req)
}

func main() {
	client.Transport = &transport{T: http.DefaultTransport}
	for {
		checkCookie()
		checkThread()
		checkPost()
		time.Sleep(time.Minute)
	}
}

func checkCookie() {
	// 检查cookie是否过期
	resp, err := client.Get("https://bbs.deepin.org/api/v1/user/msg/count")
	if err != nil {
		log.Println(err)
		return
	}
	resp.Body.Close()
	if resp.StatusCode != 200 {
		time.Sleep(time.Second)
		log.Panic(resp.Status)
	}
}

var _banUserPool sync.Map

// 禁言用户
func ban(id int64, reason string) {
	key := fmt.Sprintf("u_%d", id)
	if _, ok := _banUserPool.Load(key); ok {
		return
	}
	defer func() {
		_banUserPool.Store(key, struct{}{})
	}()
	info, err := getUserInfo(id)
	log.Printf("因%s，禁言用户：%s(%d)", reason, info.Nickname, info.ID)

	var body struct {
		UserID     int64  `json:"user_id"`
		Action     int    `json:"action"`
		BeginAt    string `json:"begin_at"`
		Reason     string `json:"reason"`
		HideThread bool   `json:"hide_thread"`
		Admin      string `json:"admin"`
	}
	body.UserID = id
	body.Action = 2
	body.BeginAt = "2022-05-19 09:35:10"
	body.Reason = "「恶意灌水机器人」：" + reason
	body.HideThread = true
	body.Admin = "bot"
	data, err := json.Marshal(&body)
	if err != nil {
		log.Fatal(err)
	}
	resp, err := client.Post("https://bbs.deepin.org/api/v1/user/crime", "application/json", bytes.NewReader(data))
	if err != nil {
		log.Println(err)
		return
	}
	defer resp.Body.Close()
	data, err = io.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return
	}
	log.Println(resp.Status, string(data))
}
