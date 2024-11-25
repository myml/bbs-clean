package main

import (
	"encoding/json"
	"io"
	"log"
	"strings"
	"time"
)

// 检查最新回复贴
func checkPost() {
	resp, err := client.Get("https://bbs.deepin.org/api/v2/public/posts?offset=0&limit=30")
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
	var result []struct {
		ID         int       `json:"id"`
		ThreadID   int       `json:"thread_id"`
		ForumID    int       `json:"forum_id"`
		UserID     int64     `json:"user_id"`
		IsFirst    int       `json:"is_first"`
		ImagesNum  int       `json:"images_num"`
		FilesNum   int       `json:"files_num"`
		Message    string    `json:"message"`
		MessageFmt string    `json:"message_fmt"`
		LikeCnt    int       `json:"like_cnt"`
		CreatedAt  time.Time `json:"created_at"`
		UpdatedAt  time.Time `json:"updated_at"`
	}
	err = json.Unmarshal(data, &result)
	if err != nil {
		log.Println(err)
		return
	}
	postCount := make(map[int64]int)
	for i := range result {
		info, err := getUserInfo(result[i].UserID)
		if err != nil {
			log.Println("get user info: %w", err)
			continue
		}
		if !info.AllowSpeak {
			continue
		}
		// 跳过高等级用户
		if info.Levels.ID > 2 {
			continue
		}
		log.Println(info.Nickname, "发布回复：", truncation(result[i].MessageFmt))
		// 一半以上的回复为同一个人发表，判断是在刷广告
		postCount[info.ID]++
		if postCount[info.ID] > len(result)/2 {
			ban(info.ID, "因账户短时间发帖过多")
			continue
		}
		// 内容的链接数超过100个，认为是在恶意发布广告
		linkCount := strings.Count(result[i].Message, `href="http`)
		if linkCount > 100 {
			ban(result[i].UserID, "因贴子链接数过多")
			continue
		}
		if linkCount > 0 && len(envAI) > 0 {
			// 通过机器学习判断是否是广告
			is, err := isAd(result[i].MessageFmt)
			if err != nil {
				log.Println("is ad:", err)
				return
			}
			if is {
				ban(result[i].UserID, "因机器学习判断")
				continue
			}
		}
		if linkCount > 0 && len(envHGToken) > 0 {
			// 通过机器学习判断是否是广告
			is, err := huggingfaceAPIAD(result[i].MessageFmt)
			if err != nil {
				log.Println("hg is ad:", err)
				return
			}
			if is {
				ban(result[i].UserID, "因机器学习判断")
				continue
			}
		}
	}
}
