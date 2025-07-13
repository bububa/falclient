package storage

import "io"

type UploadRequest struct {
	Filename    string    `json:"filename,omitempty"`
	ContentType string    `json:"content_type,omitempty"`
	Reader      io.Reader `json:"-"`
}

type UploadPartRequest struct {
	CreateUploadResult
	ContentType string    `json:"content_type,omitempty"`
	PartNumber  int       `json:"part_number,omitempty"`
	Reader      io.Reader `json:"-"`
}

type CompleteUploadRequest struct {
	CreateUploadResult
	Parts []UploadPart `json:"parts,omitempty"`
}
