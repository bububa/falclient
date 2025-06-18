package queue

import (
	"encoding/json"
)

type Status struct {
	RequestID     string     `json:"request_id,omitempty"`
	ResponseURL   string     `json:"response_url,omitempty"`
	StatusURL     string     `json:"status_url,omitempty"`
	CancelURL     string     `json:"cancel_url,omitempty"`
	Status        StatusType `json:"status,omitempty"`
	QueuePosition int        `json:"queue_position,omitempty"`
	Logs          []Log      `json:"logs,omitempty"`
	Metrics       *Metrics   `json:"metrics,omitempty"`
}

type StatusType string

const (
	IN_QUEUE               StatusType = "IN_QUEUE"
	IN_PROGRESS            StatusType = "IN_PROGRESS"
	COMPLETED              StatusType = "COMPLETED"
	CANCELLATION_REQUESTED StatusType = "CANCELLATION_REQUESTED"
	ALREADY_COMPLETED      StatusType = "ALREADY_COMPLETED"
)

type LogLevel string

const (
	STDERR LogLevel = "STDERR"
	STDOUT LogLevel = "STDOUT"
	ERROR  LogLevel = "ERROR"
	INFO   LogLevel = "INFO"
	WARN   LogLevel = "WARN"
	DEBUG  LogLevel = "DEBUG"
)

type Log struct {
	Message   string   `json:"message,omitempty"`
	Level     LogLevel `json:"level,omitempty"`
	Source    string   `json:"source,omitempty"`
	Timestamp string   `json:"timestamp,omitempty"`
}

type Metrics struct {
	InferenceTime float64 `json:"inference_time,omitempty"`
}

type CancelResponse struct {
	Status StatusType `json:"status,omitempty"`
}

type Callback func(*Status)

type QueueMode int

const (
	POLL QueueMode = iota
	STREAM
)

type SubmitRequest struct {
	Mode       QueueMode `json:"mode,omitempty"`
	Input      any       `json:"input,omitempty"`
	Callback   Callback  `json:"-"`
	WebhookURL string    `json:"-"`
}

type SubmitOption func(*SubmitRequest)

func WithInput(v any) SubmitOption {
	return func(r *SubmitRequest) {
		r.Input = v
	}
}

func WithCallback(cb Callback) SubmitOption {
	return func(r *SubmitRequest) {
		r.Callback = cb
	}
}

func WithWebhook(webhookURL string) SubmitOption {
	return func(r *SubmitRequest) {
		r.WebhookURL = webhookURL
	}
}

func WithMode(mode QueueMode) SubmitOption {
	return func(r *SubmitRequest) {
		r.Mode = mode
	}
}

type WebsocketEventType string

const (
	WSStart WebsocketEventType = "start"
	WSEnd   WebsocketEventType = "end"
	WSError WebsocketEventType = "error"
)

type WebsocketEvent struct {
	Type                   WebsocketEventType `json:"type,omitempty"`
	RequestID              string             `json:"request_id,omitempty"`
	Status                 int                `json:"status,omitempty"`
	Headers                map[string]string  `json:"headers,omitempty"`
	TimeToFirstByteSeconds float64            `json:"time_to_first_byte_seconds,omitempty"`
	Data                   json.RawMessage    `json:"data,omitempty"`
	Err                    error              `json:"err,omitempty"`
}
