package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

func huggingfaceAPIAD(text string) (bool, error) {
	data, err := json.Marshal(map[string]string{"inputs": text})
	if err != nil {
		log.Panic(err)
	}
	req, err := http.NewRequest(http.MethodPost, "https://api-inference.huggingface.co/models/myml/bbs-ad-en", bytes.NewReader(data))
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
