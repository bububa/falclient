package queue

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

func (q *Queue) Submit(ctx context.Context, endpoint string, opts ...SubmitOption) (string, error) {
	var req SubmitRequest
	for _, opt := range opts {
		opt(&req)
	}
	gw := fmt.Sprintf("%s/%s", QueueBaseURL, endpoint)
	if req.WebhookURL != "" {
		gw = fmt.Sprintf("%s?logs=1&&fal_webhook=", req.WebhookURL)
	}
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(req.Input); err != nil {
		return "", err
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, gw, &buf)
	if err != nil {
		return "", err
	}
	var resp Status
	if err := q.fetch(ctx, httpReq, &resp); err != nil {
		return "", err
	}
	requestID := resp.RequestID
	if req.Callback == nil && req.WebhookURL == "" {
		return requestID, nil
	}
	if req.Mode == STREAM {
		ch, err := q.Stream(ctx, endpoint, requestID)
		if err != nil {
			return requestID, err
		}
		for ev := range ch {
			if cb := req.Callback; cb != nil {
				cb(&ev)
			}
		}
		return requestID, nil
	}
	for {
		status, err := q.Status(ctx, endpoint, requestID)
		if err != nil {
			return requestID, err
		}
		if cb := req.Callback; cb != nil {
			cb(status)
		}
		if status.Status == COMPLETED {
			break
		}
		time.Sleep(5 * time.Second)
	}
	return requestID, nil
}
