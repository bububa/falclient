package storage

import "fmt"

const (
	RestAPIURL              = "https://rest.alpha.fal.ai"
	CDNURL                  = "https://v3.fal.media"
	UserAgent               = "fal-client/0.2.2 (golang)"
	MultipartThreshold      = 100 * 1024 * 1024
	MultipartChunkSize      = 10 * 1024 * 1024
	MultipartMaxConcurrency = 1
)

var (
	TokenStoreURL = fmt.Sprintf("%s/storage/auth/token?storage_type=fal-cdn-v3", RestAPIURL)
	FileUploadURL = fmt.Sprintf("%s/files/upload", CDNURL)
)
