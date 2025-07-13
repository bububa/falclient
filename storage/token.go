package storage

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

var (
	ErrTokenExpired       = errors.New("token expired")
	ErrTokenNotFound      = errors.New("token not found")
	ErrRefreshTokenFailed = errors.New("refresh token failed")
)

const tokenTimeLayout = "2006-01-02T15:04:05.999999-07:00"

type TokenTime time.Time

func (t *TokenTime) UnmarshalJSON(b []byte) error {
	// Remove quotes from JSON string
	s := string(b)
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		s = s[1 : len(s)-1]
	}

	v, err := time.Parse(tokenTimeLayout, s)
	if err != nil {
		return err
	}
	*t = TokenTime(v)
	return nil
}

func (t TokenTime) MarshalJSON() ([]byte, error) {
	v := time.Time(t)
	return json.Marshal(v.Format(tokenTimeLayout))
}

func (t TokenTime) String() string {
	return time.Time(t).String()
}

func (t TokenTime) Time() time.Time {
	return time.Time(t)
}

func (t TokenTime) IsZero() bool {
	return t.Time().IsZero()
}

type Token struct {
	Token     string    `json:"token,omitempty"`
	TokenType string    `json:"token_type,omitempty"`
	BaseURL   string    `json:"base_url,omitempty"`
	CreatedAt TokenTime `json:"created_at,omitzero"`
	ExpireAt  TokenTime `json:"expires_at,omitzero"`
}

func (t Token) Expired() bool {
	return t.ExpireAt.Time().Before(time.Now())
}

type TokenStore interface {
	Get(context.Context, *Token) error
	Set(context.Context, *Token) error
}

type MemoryTokenStore struct {
	token *Token
}

func (s *MemoryTokenStore) Get(ctx context.Context, token *Token) error {
	if s.token == nil {
		return ErrTokenNotFound
	} else if s.token.Expired() {
		return ErrTokenExpired
	}
	*token = *s.token
	return nil
}

func (s *MemoryTokenStore) Set(ctx context.Context, token *Token) error {
	s.token = token
	return nil
}

type TokenManager struct {
	http  *http.Client
	key   string
	store TokenStore
}

func NewTokenManager(key string, store TokenStore) *TokenManager {
	return &TokenManager{
		key:   key,
		store: store,
		http:  http.DefaultClient,
	}
}

func (m *TokenManager) SetHTTPClient(clt *http.Client) {
	m.http = clt
}

func (m *TokenManager) Refresh(ctx context.Context, token *Token) error {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, TokenStoreURL, bytes.NewReader([]byte("{}")))
	if err != nil {
		return fmt.Errorf("refresh token failed: %w", err)
	}
	httpReq.Header.Set("Authorization", fmt.Sprintf("Key %s", m.key))
	httpReq.Header.Set("Accept", "application/json")
	httpReq.Header.Set("Content-Type", "application/json")
	httpResp, err := m.http.Do(httpReq)
	if err != nil {
		return errors.Join(ErrRefreshTokenFailed, err)
	}
	defer httpResp.Body.Close()
	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		body, _ := io.ReadAll(httpResp.Body)
		return fmt.Errorf("%w, %s", ErrRefreshTokenFailed, string(body))
	}
	if err := json.NewDecoder(httpResp.Body).Decode(token); err != nil {
		return fmt.Errorf("refresh token failed: %w", err)
	}
	return m.store.Set(ctx, token)
}

func (m *TokenManager) Token(ctx context.Context, token *Token) error {
	if err := m.store.Get(ctx, token); err != nil {
		if errors.Is(err, ErrTokenNotFound) || errors.Is(err, ErrTokenExpired) {
			return m.Refresh(ctx, token)
		}
		return fmt.Errorf("get token failed:%w", err)
	}
	return nil
}
