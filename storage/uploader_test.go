package storage

import (
	"context"
	"os"
	"testing"
)

func TestUploader(t *testing.T) {
	store := new(MemoryTokenStore)
	u := NewUploader(os.Getenv("FAL_KEY"), store)
	fname := os.Getenv("FILE")
	fp, err := os.Open(fname)
	if err != nil {
		t.Error(err)
		return
	}
	defer fp.Close()
	req := UploadRequest{
		Filename: fname,
		Reader:   fp,
	}
	accessURL, err := u.Upload(context.Background(), &req)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(accessURL)
}
