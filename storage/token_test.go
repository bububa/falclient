package storage

import (
	"context"
	"encoding/json"
	"os"
	"testing"
)

func TestTokenManager(t *testing.T) {
	store := new(MemoryTokenStore)
	m := NewTokenManager(os.Getenv("FAL_KEY"), store)
	var token Token
	if err := m.Token(context.Background(), &token); err != nil {
		t.Error(err)
		return
	}
	{
		bs, _ := json.MarshalIndent(token, "", "  ")
		t.Log(string(bs))
	}
}
