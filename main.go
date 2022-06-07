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

var cookie = os.Getenv("cookie")

var client http.Client

type transport struct {
	T http.RoundTripper
}

func (t *transport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Add("cookie", cookie)
	return t.T.RoundTrip(req)
}

func main() {
	client.Transport = &transport{T: http.DefaultTransport}
	for {
		check()
		time.Sleep(time.Minute)
	}
}

var store sync.Map

func check() {
	resp, err := client.Get("https://bbs.deepin.org/api/v1/thread/index?order=updated_at&limit=20&where=&offset=0")
	if err != nil {
		log.Println(err)
		return
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return
	}
	var result struct {
		ThreadIndex []struct {
			ID   int `json:"id"`
			Top  int `json:"top"`
			User struct {
				Level int `json:"level"`
				ID    int `json:"id"`
			} `json:"user"`
		}
	}
	err = json.Unmarshal(data, &result)
	if err != nil {
		log.Println(err)
		return
	}
	m := make(map[int]int)
	for i := range result.ThreadIndex {
		t := result.ThreadIndex[i]
		if t.Top == 1 {
			continue
		}
		if t.User.Level > 2 {
			continue
		}
		m[t.User.ID]++
		if m[t.User.ID] >= 3 {
			log.Println("因账户发帖过多，禁言用户", m[t.User.ID], t.User.ID)
			ban(t.User.ID)
			return
		}
		key := fmt.Sprintf("t_%d", t.ID)
		if _, ok := store.Load(key); ok {
			continue
		}
		store.Store(key, struct{}{})
		c, err := countHTTP(t.ID)
		if err != nil {
			log.Println(err)
			return
		}
		if c {
			log.Println("因帖子链接过多，禁言用户", t.ID, t.User.ID)
			ban(t.User.ID)
			return
		}
	}
}
func countHTTP(id int) (bool, error) {
	resp, err := http.Get(fmt.Sprintf("https://bbs.deepin.org/api/v1/thread/info?id=%d", id))
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}
	log.Println(id, bytes.Count(data, []byte("http")))
	if bytes.Count(data, []byte("http")) >= 100 {
		return true, nil
	}
	return false, nil
}
func ban(id int) {
	key := fmt.Sprintf("u_%d", id)
	if _, ok := store.Load(key); ok {
		return
	}
	defer func() {
		store.Store(key, struct{}{})
	}()
	var body struct {
		UserID     int    `json:"user_id"`
		Action     int    `json:"action"`
		BeginAt    string `json:"begin_at"`
		Reason     string `json:"reason"`
		HideThread bool   `json:"hide_thread"`
		Admin      string `json:"admin"`
	}
	body.UserID = id
	body.Action = 2
	body.BeginAt = "2022-05-19 09:35:10"
	body.Reason = "疑似恶意灌水"
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
