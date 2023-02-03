package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
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
	{
		// 刷新cookie
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
			ID      int    `json:"id"`
			Top     int    `json:"top"`
			Subject string `json:"subject"`
			User    struct {
				Level    int    `json:"level"`
				ID       int    `json:"id"`
				Nickname string `json:"nickname"`
			} `json:"user"`
		}
	}
	err = json.Unmarshal(data, &result)
	if err != nil {
		log.Println(err)
		return
	}
	log.Println(result)

	m := make(map[int]int)
	for i := range result.ThreadIndex {
		t := result.ThreadIndex[i]
		if t.Top == 1 {
			continue
		}
		if t.User.Level > 2 {
			continue
		}

		for _, keyword := range []string{"Dumps", "Casinos", "dumps", "exam"} {
			if strings.Contains(result.ThreadIndex[i].Subject, keyword) {
				log.Printf("因包含关键词(%s)，禁言用户：%s(%d)", keyword, t.User.Nickname, t.User.ID)
				ban(t.User.ID, "因发帖包含关键词 "+keyword)
				return
			}
		}

		m[t.User.ID]++
		log.Printf("用户：%s 帖子数：%d", t.User.Nickname, m[t.User.ID])
		if m[t.User.ID] > 2 {
			threadsCount, err := countThread(t.User.ID)
			if err != nil {
				log.Println(err)
				return
			}
			if threadsCount > 5 {
				return
			}
			log.Printf("因账户发帖过多(%d个)，禁言用户：%s(%d)", m[t.User.ID], t.User.Nickname, t.User.ID)
			ban(t.User.ID, "因账户短时间发帖过多")
			return
		}

		key := fmt.Sprintf("t_%d", t.ID)
		if _, ok := store.Load(key); ok {
			continue
		}
		store.Store(key, struct{}{})

		linksCount, err := countHTTP(t.ID)
		if err != nil {
			log.Println(err)
			return
		}

		log.Printf("帖子：https://bbs.deepin.org/post/%d 连接数：%d", t.ID, linksCount)

		if linksCount >= 100 {
			log.Printf("因帖子%d链接过多，禁言用户: %s", t.ID, t.User.Nickname)
			ban(t.User.ID, "因贴子链接数过多")
			return
		}
	}
}
func countThread(id int) (int, error) {
	resp, err := http.Get(fmt.Sprintf("https://bbs.deepin.org/api/v1/user/info?id=%d", id))
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}
	var info struct {
		ID           int `json:"id"`
		ThreadsCount int `json:"threads_cnt"`
		PostsCount   int `json:"posts_cnt"`
	}
	err = json.Unmarshal(data, &info)
	if err != nil {
		return 0, err
	}
	return info.ThreadsCount, nil
}
func countHTTP(id int) (int, error) {
	resp, err := http.Get(fmt.Sprintf("https://bbs.deepin.org/api/v1/thread/info?id=%d", id))
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}
	return bytes.Count(data, []byte("http")), nil
}
func ban(id int, reason string) {
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
