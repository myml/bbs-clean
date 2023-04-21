package main

import (
	"encoding/json"
	"fmt"
	"io"
	"sync"
	"time"
)

// UserInfo 用户信息
type UserInfo struct {
	ID           int64     `json:"id"`
	AccountID    int       `json:"account_id"`
	GroupID      int       `json:"group_id"`
	GroupName    string    `json:"group_name"`
	Email        string    `json:"email"`
	EmailChecked int       `json:"email_checked"`
	Username     string    `json:"username"`
	Realname     string    `json:"realname"`
	Nickname     string    `json:"nickname"`
	Mobile       string    `json:"mobile"`
	Qq           string    `json:"qq"`
	ThreadsCnt   int       `json:"threads_cnt"`
	PostsCnt     int       `json:"posts_cnt"`
	MsgCnt       int       `json:"msg_cnt"`
	CreditsNum   int       `json:"credits_num"`
	CreateIP     string    `json:"create_ip"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	LoginIP      string    `json:"login_ip"`
	LoginDate    time.Time `json:"login_date"`
	LoginsCnt    int       `json:"logins_cnt"`
	Avatar       string    `json:"avatar"`
	DigestsNum   int       `json:"digests_num"`
	State        int       `json:"state"`
	LikeCnt      int       `json:"like_cnt"`
	FavouriteCnt int       `json:"favourite_cnt"`
	AllowSpeak   bool      `json:"allow_speak"`
	Desc         string    `json:"desc"`
	Level        int       `json:"level"`
	Levels       struct {
		ID        int    `json:"id"`
		Admin     string `json:"admin"`
		ColorID   int    `json:"color_id"`
		LevelIcon string `json:"level_icon"`
		LevelName string `json:"level_name"`
		Min       int    `json:"min"`
		Max       int    `json:"max"`
		CreatedAt string `json:"created_at"`
		UpdatedAt string `json:"updated_at"`
	} `json:"levels"`
}

var _userInfoCache sync.Map

// 获取用户信息
func getUserInfo(id int64) (*UserInfo, error) {
	if v, ok := _userInfoCache.Load(id); ok {
		return v.(*UserInfo), nil
	}
	resp, err := client.Get(fmt.Sprintf("https://bbs.deepin.org/api/v1/user/info?id=%d", id))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var info UserInfo
	err = json.Unmarshal(data, &info)
	if err != nil {
		return nil, err
	}
	_userInfoCache.Store(id, &info)
	// 缓存在一个小时之后刷新
	time.AfterFunc(time.Hour, func() {
		_userInfoCache.Delete(id)
	})
	return &info, nil
}
