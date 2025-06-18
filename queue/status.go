package queue

import (
	"context"
	"fmt"
	"net/http"
)

// Status Gets the status of a request
func (q *Queue) Status(ctx context.Context, endpoint string, requestID string) (*Status, error) {
	appID, err := AppIDFromEndpoint(endpoint)
	if err != nil {
		return nil, err
	}
	gw := fmt.Sprintf("%s/%s/requests/%s/status?logs=1", QueueBaseURL, appID.URLString(), requestID)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, gw, nil)
	if err != nil {
		return nil, err
	}
	var resp Status
	if err := q.fetch(ctx, httpReq, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
