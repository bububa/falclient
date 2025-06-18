package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"
)

func TestSubmitPoll(t *testing.T) {
	ctx := context.Background()
	key := os.Getenv("KEY")
	prompt := os.Getenv("PROMPT")
	if key == "" || prompt == "" {
		t.Error("missing key/prompt")
		return
	}
	endpoint := "fal-ai/flux-pro/kontext/text-to-image"
	queue := NewQueue(key)
	reqID, err := queue.Submit(ctx, endpoint, WithInput(map[string]string{
		"prompt": prompt,
	}), WithCallback(func(v *Status) {
		bs, _ := json.MarshalIndent(v, "", "  ")
		fmt.Println(string(bs))
	}))
	if err != nil {
		t.Error(err)
		return
	}
	t.Logf("RequestID: %s", reqID)
	var resp struct {
		Images []struct {
			URL         string `json:"url,omitempty"`
			Width       int    `json:"width,omitempty"`
			Height      int    `json:"height,omitempty"`
			ContentType string `json:"content_type,omitempty"`
		} `json:"images,omitempty"`
		Seed            int64  `json:"seed,omitempty"`
		HasNSFWConcepts []bool `json:"has_nsfw_concepts,omitempty"`
		Prompt          string `json:"prompt,omitempty"`
	}
	if err := queue.Response(ctx, endpoint, reqID, &resp); err != nil {
		t.Error(err)
		return
	}
	bs, _ := json.MarshalIndent(resp, "", "  ")
	t.Log(string(bs))
}

func TestSubmitStream(t *testing.T) {
	ctx := context.Background()
	key := os.Getenv("KEY")
	prompt := os.Getenv("PROMPT")
	if key == "" || prompt == "" {
		t.Error("missing key/prompt")
		return
	}
	endpoint := "fal-ai/flux-pro/kontext/text-to-image"
	queue := NewQueue(key)
	reqID, err := queue.Submit(ctx, endpoint, WithInput(map[string]string{
		"prompt": prompt,
	}), WithCallback(func(v *Status) {
		bs, _ := json.MarshalIndent(v, "", "  ")
		fmt.Println(string(bs))
	}), WithMode(STREAM))
	if err != nil {
		t.Error(err)
		return
	}
	var resp struct {
		Images []struct {
			URL         string `json:"url,omitempty"`
			Width       int    `json:"width,omitempty"`
			Height      int    `json:"height,omitempty"`
			ContentType string `json:"content_type,omitempty"`
		} `json:"images,omitempty"`
		Seed            int64  `json:"seed,omitempty"`
		HasNSFWConcepts []bool `json:"has_nsfw_concepts,omitempty"`
		Prompt          string `json:"prompt,omitempty"`
	}
	if err := queue.Response(ctx, endpoint, reqID, &resp); err != nil {
		t.Error(err)
		return
	}
	bs, _ := json.MarshalIndent(resp, "", "  ")
	t.Log(string(bs))
}
