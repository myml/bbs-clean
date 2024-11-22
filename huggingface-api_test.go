package main

import (
	"log"
	"testing"
)

func Test_huggingfaceAPIAD(t *testing.T) {
	text := `I usually turn on eye protection mode and it's great to know that you guys have added a shortcut key for it. basketball stars It would be much more convenient and time-saving than if I had to turn it on manually.`
	isAD, err := huggingfaceAPIAD(text)
	if err != nil {
		log.Fatal(err)
	}
	t.Log(isAD, err)
}
