package webhook

import "encoding/json"

type Status string

const (
	OK    Status = "OK"
	ERROR Status = "ERROR"
)

type Request struct {
	RequestID        string          `json:"request_id,omitempty"`
	GatewayRequestID string          `json:"gateway_request_id,omitempty"`
	Status           Status          `json:"status,omitempty"`
	Err              string          `json:"error,omitempty"`
	Payload          json.RawMessage `json:"payload,omitempty"`
	PayloadError     string          `json:"payload_error,omitempty"`
}

func (r Request) IsError() bool {
	return r.Status == ERROR || r.PayloadError != ""
}

func (r Request) Error() string {
	if r.Status == ERROR {
		return r.Err
	}
	if r.PayloadError != "" {
		return r.PayloadError
	}
	return ""
}
