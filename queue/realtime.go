package queue

import (
	"context"
	"errors"
	"io"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

func (q *Queue) Realtime(ctx context.Context, endpoint string, input any) (<-chan WebsocketEvent, error) {
	appID, err := AppIDFromEndpoint(endpoint)
	if err != nil {
		return nil, err
	}
	conn, err := q.WS(ctx, appID.URLString())
	if err != nil {
		return nil, err
	}
	defer conn.CloseNow()
	if err := wsjson.Write(ctx, conn, input); err != nil {
		return nil, err
	}
	ch := make(chan WebsocketEvent)
	go func() {
		defer close(ch)
		for {
			var ev WebsocketEvent
			if err := wsjson.Read(ctx, conn, &ev); err != nil {
				if errors.Is(err, io.EOF) {
					ch <- WebsocketEvent{Type: WSError, Err: err}
					break
				} else {
					conn.Close(websocket.StatusAbnormalClosure, err.Error())
				}
			}
			ch <- ev
		}
	}()
	return ch, nil
}
