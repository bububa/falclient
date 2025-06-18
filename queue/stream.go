package queue

import (
	"context"
	"fmt"
	"net/http"
)

// Stream Gets the stream status of a request
func (q *Queue) Stream(ctx context.Context, endpoint string, requestID string) (<-chan Status, error) {
	appID, err := AppIDFromEndpoint(endpoint)
	if err != nil {
		return nil, err
	}
	gw := fmt.Sprintf("%s/%s/requests/%s/status/stream?logs=1", QueueBaseURL, appID.URLString(), requestID)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, gw, nil)
	if err != nil {
		return nil, err
	}
	return q.SSE(ctx, httpReq)
}
