package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

// MLRequest gradio接口请求
type MLRequest struct {
	FnIndex     int      `json:"fn_index"`
	Data        []string `json:"data"`
	SessionHash string   `json:"session_hash"`
}

// MLResponse gradio接口返回
type MLResponse struct {
	Data []struct {
		Label       string `json:"label"`
		Confidences []struct {
			Label      string  `json:"label"` // LABEL_0 正常 LABEL_1 广告
			Confidence float64 `json:"confidence"`
		} `json:"confidences"`
	} `json:"data"`
	IsGenerating    bool    `json:"is_generating"`
	Duration        float64 `json:"duration"`
	AverageDuration float64 `json:"average_duration"`
}

// 使用机器学习判断是否违规内容，返回true则认为违规
func isAd(text string) (bool, error) {
	req := MLRequest{
		FnIndex:     0,
		Data:        []string{text},
		SessionHash: "qmphwek45ul",
	}
	data, err := json.Marshal(req)
	if err != nil {
		log.Panic(err)
	}
	resp, err := http.Post(envAI, "application/json", bytes.NewReader(data))
	if err != nil {
		return false, fmt.Errorf("can not connect server: %w", err)
	}
	defer resp.Body.Close()
	data, err = io.ReadAll(resp.Body)
	if err != nil {
		return false, fmt.Errorf("can not read body: %w", err)
	}
	var result MLResponse
	err = json.Unmarshal(data, &result)
	if err != nil {
		return false, fmt.Errorf("can not unmarshal result: %w", err)
	}
	normalConfidence := 0.0
	adConfidence := 0.0
	if len(result.Data) == 0 {
		log.Printf("can not get result: %#v\n", result)
	}
	for _, confidence := range result.Data[0].Confidences {
		switch confidence.Label {
		case "LABEL_0":
			normalConfidence = confidence.Confidence
		case "LABEL_1":
			adConfidence = confidence.Confidence
		}
	}
	if normalConfidence < 0.2 && adConfidence > 0.8 {
		return true, nil
	}
	return false, nil
}
