package queue

import (
	"context"
	"fmt"
	"net/http"
)

// Response Gets the response of a request
func (q *Queue) Response(ctx context.Context, endpoint string, requestID string, resp any) error {
	appID, err := AppIDFromEndpoint(endpoint)
	if err != nil {
		return err
	}
	gw := fmt.Sprintf("%s/%s/requests/%s", QueueBaseURL, appID.URLString(), requestID)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, gw, nil)
	if err != nil {
		return err
	}
	return q.fetch(ctx, httpReq, resp)
}
