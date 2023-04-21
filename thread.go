package main

import (
	"encoding/json"
	"io"
	"log"
	"strings"
)

// 检查首页主题贴
func checkThread() {
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
				ID       int64  `json:"id"`
				Nickname string `json:"nickname"`
			} `json:"user"`
		}
	}
	err = json.Unmarshal(data, &result)
	if err != nil {
		log.Println(err)
		return
	}

	threadCount := make(map[int64]int)
	for i := range result.ThreadIndex {
		t := result.ThreadIndex[i]
		// 跳过置顶帖
		if t.Top == 1 {
			continue
		}
		// 跳过高等级用户
		if t.User.Level > 2 {
			continue
		}
		log.Println(t.User.Nickname, "发布贴子：", t.Subject[:10]+"...")

		// 贴子标题包含关键字，认为是散发广告
		for _, keyword := range Keywords {
			if strings.Contains(strings.ToLower(result.ThreadIndex[i].Subject), keyword) {
				ban(t.User.ID, "因发帖包含关键词 "+keyword)
				return
			}
		}
		// 用户短时间发帖超过3个，并且历史发帖数少于5个，认为是新号在恶意批量发广告
		threadCount[t.User.ID]++
		if threadCount[t.User.ID] > 3 {
			info, err := getUserInfo(t.User.ID)
			if err != nil {
				log.Println(err)
				return
			}
			if info.ThreadsCnt > 5 {
				return
			}
			ban(t.User.ID, "因账户短时间发帖过多")
			return
		}
		// 通过机器学习判断是否是广告
		is, err := isAd(t.Subject)
		if err != nil {
			log.Println("is ad:", err)
			return
		}
		if is {
			ban(t.User.ID, "因机器学习判断")
			return
		}
	}
}
