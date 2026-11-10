package kakao

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const (
	contentTypeForm = "application/x-www-form-urlencoded;charset=utf-8"
	contentTypeJSON = "application/json"
)

// authKind selects the credential sent with a request.
type authKind int

const (
	authNone     authKind = iota // no Authorization header
	authRESTKey                  // Authorization: KakaoAK <REST API key>
	authAdminKey                 // Authorization: KakaoAK <admin key>
	authBearer                   // Authorization: Bearer <token>
)

// request describes one API call. At most one of form and json is set.
type request struct {
	method string
	host   Host
	path   string
	auth   authKind
	token  string // for authBearer
	query  url.Values
	form   url.Values
	json   any
	// idempotent marks a non-GET request as safe to retry.
	idempotent bool
}

// rawBody receives the undecoded response body, for non-JSON responses
// such as static map images.
type rawBody struct {
	ContentType string
	Data        []byte
}

// do sends r and decodes a successful JSON response into out. out may be
// nil to discard the body, or a *rawBody to keep it undecoded.
func (c *Client) do(ctx context.Context, r *request, out any) error {
	authz, err := c.authorization(r)
	if err != nil {
		return err
	}
	body, contentType, err := encodeBody(r)
	if err != nil {
		return err
	}
	u := c.baseURLs[r.host] + r.path
	if len(r.query) > 0 {
		u += "?" + r.query.Encode()
	}

	attempts := 1
	if c.retry != nil && (r.method == http.MethodGet || r.idempotent) {
		attempts = c.retry.MaxAttempts
	}
	for attempt := 1; ; attempt++ {
		header, err := c.send(ctx, r, u, authz, contentType, body, out)
		if err == nil {
			return nil
		}
		if attempt >= attempts || ctx.Err() != nil || !retryable(err) {
			return err
		}
		t := time.NewTimer(c.retry.backoff(attempt, header))
		select {
		case <-ctx.Done():
			t.Stop()
			return err
		case <-t.C:
		}
	}
}

// send performs a single attempt. It returns the response header, if a
// response was received, for the retry backoff.
func (c *Client) send(ctx context.Context, r *request, u, authz, contentType string, body []byte, out any) (http.Header, error) {
	var rd io.Reader
	if body != nil {
		rd = bytes.NewReader(body)
	}
	req, err := http.NewRequestWithContext(ctx, r.method, u, rd)
	if err != nil {
		return nil, fmt.Errorf("kakao: build request: %w", err)
	}
	req.Header.Set("User-Agent", c.userAgent)
	if authz != "" {
		req.Header.Set("Authorization", authz)
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(io.LimitReader(resp.Body, c.maxBodySize+1))
	if err != nil {
		return resp.Header, fmt.Errorf("kakao: read %s %s response: %w", r.method, r.path, err)
	}
	if int64(len(data)) > c.maxBodySize {
		return resp.Header, fmt.Errorf("kakao: %s %s response exceeds %d bytes", r.method, r.path, c.maxBodySize)
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return resp.Header, parseError(r.host, resp.StatusCode, data)
	}

	switch out := out.(type) {
	case nil:
	case *rawBody:
		out.ContentType = resp.Header.Get("Content-Type")
		out.Data = data
	default:
		if len(data) == 0 {
			return resp.Header, nil
		}
		if err := json.Unmarshal(data, out); err != nil {
			return resp.Header, fmt.Errorf("kakao: decode %s %s response: %w", r.method, r.path, err)
		}
	}
	return resp.Header, nil
}

func (c *Client) authorization(r *request) (string, error) {
	switch r.auth {
	case authNone:
		return "", nil
	case authRESTKey:
		if c.restKey == "" {
			return "", fmt.Errorf("%w: REST API key (use WithRESTKey)", ErrMissingCredential)
		}
		return "KakaoAK " + c.restKey, nil
	case authAdminKey:
		if c.adminKey == "" {
			return "", fmt.Errorf("%w: admin key (use WithAdminKey)", ErrMissingCredential)
		}
		return "KakaoAK " + c.adminKey, nil
	case authBearer:
		if r.token == "" {
			return "", fmt.Errorf("%w: access token", ErrMissingCredential)
		}
		return "Bearer " + r.token, nil
	default:
		return "", fmt.Errorf("kakao: unknown auth kind %d", r.auth)
	}
}

func encodeBody(r *request) (body []byte, contentType string, err error) {
	switch {
	case r.form != nil && r.json != nil:
		return nil, "", fmt.Errorf("kakao: %s %s: both form and JSON body set", r.method, r.path)
	case r.form != nil:
		return []byte(r.form.Encode()), contentTypeForm, nil
	case r.json != nil:
		b, err := json.Marshal(r.json)
		if err != nil {
			return nil, "", fmt.Errorf("kakao: encode %s %s body: %w", r.method, r.path, err)
		}
		return b, contentTypeJSON, nil
	default:
		return nil, "", nil
	}
}
