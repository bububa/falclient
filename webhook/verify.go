package webhook

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/lestrrat-go/httprc/v3"
	"github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/lestrrat-go/jwx/v3/jws"
)

var (
	cache     *jwk.Cache
	onceCache = new(sync.Once)
)

func JWKCache(ctx context.Context) (*jwk.Cache, error) {
	var retErr error
	onceCache.Do(func() {
		if c, err := jwk.NewCache(ctx, httprc.NewClient()); err != nil {
			onceCache = new(sync.Once)
			retErr = err
		} else if err := c.Register(ctx, JWSKEndpoint, jwk.WithConstantInterval(time.Hour*24)); err != nil {
			onceCache = new(sync.Once)
			retErr = err
		} else {
			cache = c
		}
	})
	return cache, retErr
}

func Verify(ctx context.Context, httpReq *http.Request, req *Request) error {
	cache, err := JWKCache(ctx)
	if err != nil {
		return err
	}
	jwkSet, err := cache.Lookup(ctx, JWSKEndpoint)
	if err != nil {
		return err
	}
	header := httpReq.Header
	timestamp := header.Get("X-Fal-Webhook-Timestamp")
	ts, _ := strconv.ParseInt(timestamp, 10, 64)
	if time.Since(time.Unix(ts, 0)).Abs().Seconds() > 300 {
		return errors.New("invalid timestamp")
	}
	signature := header.Get("X-Fal-Webhook-Signature")
	sig, err := hex.DecodeString(signature)
	if err != nil {
		return err
	}
	requestID := header.Get("X-Fal-Webhook-Request-Id")
	userID := header.Get("X-Fal-Webhook-User-Id")
	var verifyBuf bytes.Buffer
	verifyBuf.WriteString(requestID)
	verifyBuf.WriteByte('\n')
	verifyBuf.WriteString(userID)
	verifyBuf.WriteByte('\n')
	verifyBuf.WriteString(timestamp)
	verifyBuf.WriteByte('\n')
	var reqBuf bytes.Buffer
	tee := io.TeeReader(httpReq.Body, &reqBuf)
	h := sha256.New()
	io.Copy(h, tee)
	hashBody := hex.EncodeToString(h.Sum(nil))
	verifyBuf.WriteString(hashBody)
	if _, err := jws.Verify(sig, jws.WithKeySet(jwkSet)); err != nil {
		return err
	}
	return json.NewDecoder(&reqBuf).Decode(req)
}
