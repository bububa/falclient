package storage

type CreateUploadResult struct {
	AccessURL       string `json:"access_url,omitempty"`
	UploadID        string `json:"upload_id,omitempty"`
	UploadSignature string `json:"upload_signature,omitempty"`
}

type UploadPart struct {
	PartNumber int    `json:"partNumber,omitempty"`
	ETag       string `json:"etag,omitempty"`
}
