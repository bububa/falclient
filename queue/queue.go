package queue

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/coder/websocket"
	"github.com/tmaxmax/go-sse"
)

type QueueOption func(*Queue)

func WithDebug(v bool) QueueOption {
	return func(q *Queue) {
		q.debug = v
	}
}

func WithHttpClient(clt *http.Client) QueueOption {
	return func(q *Queue) {
		q.http = clt
	}
}

type Queue struct {
	// token authorized key
	token string
	http  *http.Client
	debug bool
}

func NewQueue(token string, opts ...QueueOption) *Queue {
	ret := &Queue{token: token}
	for _, opt := range opts {
		opt(ret)
	}
	if ret.http == nil {
		ret.http = http.DefaultClient
	}
	return ret
}

func (q *Queue) SetDebug(v bool) {
	q.debug = v
}

func (q *Queue) fetch(_ context.Context, req *http.Request, resp any) error {
	req.Header.Set("Authorization", fmt.Sprintf("Key %s", q.token))
	req.Header.Set("Content-Type", "application/json")
	httpResp, err := q.http.Do(req)
	if err != nil {
		return err
	}
	defer httpResp.Body.Close()
	if httpResp.StatusCode < 200 && httpResp.StatusCode >= 300 {
		msg, _ := io.ReadAll(httpResp.Body)
		return errors.New(string(msg))
	}
	// bs, _ := io.ReadAll(httpResp.Body)
	// fmt.Println(string(bs))
	// return json.Unmarshal(bs, resp)
	return json.NewDecoder(httpResp.Body).Decode(resp)
}

func (q *Queue) SSE(ctx context.Context, req *http.Request) (<-chan Status, error) {
	req.Header.Set("Authorization", fmt.Sprintf("Key %s", q.token))
	req.Header.Set("Content-Type", "application/json")
	httpResp, err := q.http.Do(req)
	if err != nil {
		return nil, err
	}
	ch := make(chan Status)
	go func() {
		defer httpResp.Body.Close()
		defer close(ch)
		for ev, err := range sse.Read(httpResp.Body, nil) {
			if err != nil {
				// handle read error
				break // can end the loop as Read stops on first error anyway
			}
			if data := ev.Data; data != "" {
				var item Status
				if err := json.Unmarshal([]byte(data), &item); err == nil {
					ch <- item
				}
			}
		}
	}()
	return ch, nil
}

func (q *Queue) WS(ctx context.Context, appID string) (*websocket.Conn, error) {
	header := make(http.Header)
	header.Set("Authorization", fmt.Sprintf("Key %s", q.token))
	conn, _, err := websocket.Dial(ctx, fmt.Sprintf("%s/%s", WSBaseURL, appID), &websocket.DialOptions{
		HTTPClient: q.http,
		HTTPHeader: header,
	})
	if err != nil {
		return nil, err
	}
	return conn, nil
}
