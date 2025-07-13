package storage

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"strings"
	"sync"
)

var (
	ErrCreateUpload   = fmt.Errorf("create upload failed")
	ErrUploadPart     = fmt.Errorf("upload part failed")
	ErrUploadComplete = fmt.Errorf("upload complete failed")
	ErrUploadFile     = fmt.Errorf("upload file failed")
)

type Option func(*Uploader)

func WithHTTPClient(clt *http.Client) Option {
	return func(u *Uploader) {
		u.http = clt
	}
}

func WithChunkSize(size int64) Option {
	return func(u *Uploader) {
		u.chunkSize = size
	}
}

func WithThreads(n int) Option {
	return func(u *Uploader) {
		u.threads = n
	}
}

type Uploader struct {
	http         *http.Client
	tokenManager *TokenManager
	chunkSize    int64
	threads      int
}

func NewUploader(key string, store TokenStore, opts ...Option) *Uploader {
	ret := &Uploader{
		http:         http.DefaultClient,
		tokenManager: NewTokenManager(key, store),
		chunkSize:    MultipartChunkSize,
		threads:      MultipartMaxConcurrency,
	}
	for _, opt := range opts {
		opt(ret)
	}
	ret.tokenManager.SetHTTPClient(ret.http)
	return ret
}

func (u *Uploader) appendAuthHeader(req *http.Request, token *Token) {
	if token != nil {
		req.Header.Set("Authorization", fmt.Sprintf("%s %s", token.TokenType, token.Token))
	}
	req.Header.Set("User-Agent", UserAgent)
}

func (u *Uploader) create(ctx context.Context, req *UploadRequest, resp *CreateUploadResult) error {
	var token Token
	if err := u.tokenManager.Token(ctx, &token); err != nil {
		return err
	}
	gw := fmt.Sprintf("%s/files/upload/multipart", token.BaseURL)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, gw, strings.NewReader("{}"))
	if err != nil {
		return errors.Join(ErrCreateUpload, err)
	}
	u.appendAuthHeader(httpReq, &token)
	httpReq.Header.Set("Accept", "application/json")
	contentType := req.ContentType
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	httpReq.Header.Set("Content-Type", contentType)
	httpReq.Header.Set("X-Fal-File-Name", req.Filename)

	if err := u.fetch(httpReq, resp); err != nil {
		return errors.Join(ErrCreateUpload, err)
	}
	return nil
}

func (u *Uploader) uploadPart(ctx context.Context, req *UploadPartRequest, ret *UploadPart) error {
	var token Token
	if err := u.tokenManager.Token(ctx, &token); err != nil {
		return err
	}
	gw := fmt.Sprintf("%s/multipart/%s/%d", req.AccessURL, req.UploadID, req.PartNumber)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPut, gw, req.Reader)
	if err != nil {
		return errors.Join(ErrUploadPart, err)
	}
	u.appendAuthHeader(httpReq, &token)
	httpReq.Header.Set("Accept", "application/json")
	contentType := req.ContentType
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	httpReq.Header.Set("Content-Type", contentType)
	httpReq.Header.Set("Accept-Encoding", "identity") // Keep this to ensure we get ETag headers
	httpResp, err := u.http.Do(httpReq)
	if err != nil {
		return errors.Join(ErrUploadPart, err)
	}
	defer httpResp.Body.Close()
	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		body, _ := io.ReadAll(httpResp.Body)
		return fmt.Errorf("code: %d, body: %s", httpResp.StatusCode, string(body))
	}
	etag := httpResp.Header.Get("ETag")
	ret.PartNumber = req.PartNumber
	ret.ETag = etag
	return nil
}

func (u *Uploader) complete(ctx context.Context, req *CompleteUploadRequest) error {
	var token Token
	if err := u.tokenManager.Token(ctx, &token); err != nil {
		return err
	}
	gw := fmt.Sprintf("%s/multipart/%s/complete", req.AccessURL, req.UploadID)
	var buf bytes.Buffer
	payload := struct {
		Parts []UploadPart `json:"parts"`
	}{Parts: req.Parts}
	if err := json.NewEncoder(&buf).Encode(payload); err != nil {
		return errors.Join(ErrUploadComplete, err)
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, gw, &buf)
	if err != nil {
		return errors.Join(ErrUploadPart, err)
	}
	u.appendAuthHeader(httpReq, &token)
	if err := u.fetch(httpReq, nil); err != nil {
		return errors.Join(ErrUploadComplete, err)
	}
	return nil
}

func (u *Uploader) uploadFile(ctx context.Context, req *UploadRequest) (string, error) {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, FileUploadURL, req.Reader)
	if err != nil {
		return "", errors.Join(ErrUploadFile, err)
	}
	var token Token
	if err := u.tokenManager.Token(ctx, &token); err != nil {
		return "", err
	}
	u.appendAuthHeader(httpReq, &token)
	httpReq.Header.Set("X-Fal-File-Name", req.Filename)
	httpReq.Header.Set("Content-Type", req.ContentType)
	httpResp, err := u.http.Do(httpReq)
	if err != nil {
		return "", errors.Join(ErrUploadFile, err)
	}
	defer httpResp.Body.Close()
	var ret CreateUploadResult
	if err := json.NewDecoder(httpResp.Body).Decode(&ret); err != nil {
		return "", errors.Join(ErrUploadFile, err)
	}
	return ret.AccessURL, nil
}

func (u *Uploader) Upload(ctx context.Context, req *UploadRequest) (string, error) {
	if req.Filename == "" {
		req.Filename = "upload.bin"
	}
	if req.ContentType == "" {
		req.ContentType = "application/octet-stream"
	}
	var buf bytes.Buffer
	size, err := io.Copy(&buf, req.Reader)
	if err != nil {
		return "", err
	}
	if size <= MultipartThreshold {
		uploadReq := UploadRequest{
			Filename:    req.Filename,
			ContentType: req.ContentType,
			Reader:      bytes.NewReader(buf.Bytes()),
		}
		return u.uploadFile(ctx, &uploadReq)
	}
	var createResp CreateUploadResult
	if err := u.create(ctx, req, &createResp); err != nil {
		return "", err
	}
	chunks := int64(math.Ceil(float64(size) / float64(u.chunkSize)))
	bs := buf.Bytes()
	var (
		semaphore = make(chan struct{}, u.threads)
		wg        sync.WaitGroup
		partErr   error
		lock      = new(sync.Mutex)
		parts     = make([]UploadPart, chunks)
	)
	for idx := range chunks {
		startPart := idx * u.chunkSize
		endPart := min(startPart+u.chunkSize, size)
		r := bytes.NewReader(bs[startPart:endPart])
		partReq := UploadPartRequest{
			CreateUploadResult: createResp,
			ContentType:        req.ContentType,
			PartNumber:         int(idx + 1),
			Reader:             r,
		}
		wg.Add(1)
		semaphore <- struct{}{}
		go func(partReq *UploadPartRequest) {
			defer wg.Done()
			defer func() { <-semaphore }()
			var partRet UploadPart
			if err := u.uploadPart(ctx, partReq, &partRet); err != nil {
				if partErr != nil {
					partErr = err
				} else {
					partErr = errors.Join(partErr, err)
				}
				return
			}
			lock.Lock()
			parts = append(parts, partRet)
			lock.Unlock()
		}(&partReq)
	}
	wg.Wait()
	completeReq := CompleteUploadRequest{
		CreateUploadResult: createResp,
		Parts:              parts,
	}
	if err := u.complete(ctx, &completeReq); err != nil {
		return "", err
	}
	return createResp.AccessURL, nil
}

func (u *Uploader) fetch(req *http.Request, resp any) error {
	httpResp, err := u.http.Do(req)
	if err != nil {
		return err
	}
	defer httpResp.Body.Close()
	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		body, _ := io.ReadAll(httpResp.Body)
		return fmt.Errorf("code: %d, body: %s", httpResp.StatusCode, string(body))
	}
	if resp != nil {
		return json.NewDecoder(httpResp.Body).Decode(resp)
	}
	return nil
}
