package queue

import (
	"context"
	"fmt"
	"net/http"
)

// Cancel cancel a request
func (q *Queue) Cancel(ctx context.Context, endpoint string, requestID string) (StatusType, error) {
	appID, err := AppIDFromEndpoint(endpoint)
	if err != nil {
		return "", err
	}
	gw := fmt.Sprintf("%s/%s/requests/%s/cancel", QueueBaseURL, appID.URLString(), requestID)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPut, gw, nil)
	if err != nil {
		return "", err
	}
	var resp CancelResponse
	if err := q.fetch(ctx, httpReq, &resp); err != nil {
		return resp.Status, err
	}
	return resp.Status, nil
}
