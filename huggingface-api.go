package main

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
	"unicode"
)

var cache = make(map[string]*bool)

func containsChinese(s string) bool {
	for _, r := range s {
		// 判断字符是否在中文的 Unicode 范围
		if unicode.Is(unicode.Han, r) {
			return true
		}
	}
	return false
}

func huggingfaceAPIAD(text string) (bool, error) {
	d := md5.Sum([]byte(text))
	key := hex.EncodeToString(d[:])
	if cache[key] != nil {
		return *cache[key], nil
	}
	// TODO 中文模型不理想，暂不使用
	if containsChinese(text) {
		return false, nil
	}
	v, err := huggingfaceAPIADWithoutCache(text, "myml/bbs-ad-en")
	if err != nil {
		return false, err
	}
	cache[key] = &v
	time.AfterFunc(time.Hour*24, func() {
		delete(cache, key)
	})
	return v, nil
}

func huggingfaceAPIADWithoutCache(text string, module string) (bool, error) {
	data, err := json.Marshal(map[string]string{"inputs": text})
	if err != nil {
		log.Panic(err)
	}
	req, err := http.NewRequest(
		http.MethodPost,
		"https://api-inference.huggingface.co/models/"+module,
		bytes.NewReader(data),
	)
	if err != nil {
		return false, fmt.Errorf("new request: %w", err)
	}
	req.Header.Add("Authorization", "Bearer "+envHGToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, fmt.Errorf("can not connect server: %w", err)
	}
	defer resp.Body.Close()
	data, err = io.ReadAll(resp.Body)
	if err != nil {
		return false, fmt.Errorf("can not read body: %w", err)
	}

	type HuggingfaceAPIResp struct {
		Label string
		Score float64
	}
	var result [][]HuggingfaceAPIResp
	err = json.Unmarshal(data, &result)
	if err != nil {
		return false, fmt.Errorf("can not unmarshal result: %w %s", err, data)
	}
	if len(result) == 0 || len(result[0]) == 0 {
		return false, fmt.Errorf("invalid result: %s", data)
	}
	normalConfidence := 0.0
	adConfidence := 0.0
	for _, confidence := range result[0] {
		switch confidence.Label {
		case "LABEL_0":
			normalConfidence = confidence.Score
		case "LABEL_1":
			adConfidence = confidence.Score
		}
	}
	if normalConfidence <= 0.3 && adConfidence >= 0.7 {
		return true, nil
	}
	log.Printf("%s: normalConfidence: %v, adConfidence: %v\n", text[:50], normalConfidence, adConfidence)
	return false, nil
}
