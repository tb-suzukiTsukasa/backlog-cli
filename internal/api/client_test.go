package api

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

// mockHTTPClient は HTTP レスポンスを固定値で返すテスト用クライアント。
type mockHTTPClient struct {
	statusCode int
	body       string
}

func (m *mockHTTPClient) Do(_ *http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: m.statusCode,
		Body:       io.NopCloser(bytes.NewBufferString(m.body)),
		Header:     make(http.Header),
	}, nil
}

func newMockClient(t *testing.T, statusCode int, body string) *Client {
	t.Helper()
	mock := &mockHTTPClient{statusCode: statusCode, body: body}
	return &Client{host: "example.backlog.com", apiKey: "test-key", httpClient: mock}
}

func TestNewClient(t *testing.T) {
	c, err := NewClient("example.backlog.com", "test-api-key")
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	if c == nil {
		t.Fatal("NewClient() returned nil client")
	}
}

// multiMockHTTPClient は呼び出し順に異なるレスポンスを返すテスト用クライアント。
type multiMockHTTPClient struct {
	responses []struct {
		statusCode int
		body       string
	}
	index int
}

func (m *multiMockHTTPClient) Do(_ *http.Request) (*http.Response, error) {
	idx := m.index
	if idx >= len(m.responses) {
		idx = len(m.responses) - 1
	}
	m.index++
	r := m.responses[idx]
	return &http.Response{
		StatusCode: r.statusCode,
		Body:       io.NopCloser(bytes.NewBufferString(r.body)),
		Header:     make(http.Header),
	}, nil
}

func TestNewClient_EmptyHost(t *testing.T) {
	_, err := NewClient("", "key")
	if err == nil {
		t.Fatal("expected error for empty host, got nil")
	}
}

func TestNewClient_EmptyAPIKey(t *testing.T) {
	_, err := NewClient("example.backlog.com", "")
	if err == nil {
		t.Fatal("expected error for empty API key, got nil")
	}
}

func TestConvertError(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		body       string
		wantSubstr string
	}{
		{"401 no detail", 401, `{}`, "認証エラー"},
		{"401 with detail", 401, `{"errors":[{"message":"Invalid API key."}]}`, "Invalid API key."},
		{"403", 403, `{}`, "権限エラー"},
		{"404", 404, `{}`, "見つかりません"},
		{"404 with detail", 404, `{"errors":[{"message":"No such project."}]}`, "No such project."},
		{"429", 429, `{}`, "レート制限"},
		{"500 raw", 500, `internal error`, "internal error"},
		{"500 with json", 500, `{"errors":[{"message":"Server error"}]}`, "Server error"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := convertError(tt.statusCode, []byte(tt.body))
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !strings.Contains(err.Error(), tt.wantSubstr) {
				t.Errorf("error %q does not contain %q", err.Error(), tt.wantSubstr)
			}
		})
	}
}

func TestListUsers(t *testing.T) {
	body := `[{"id":1,"name":"Alice","mailAddress":"alice@example.com"},{"id":2,"name":"Bob","mailAddress":"bob@example.com"}]`
	c := newMockClient(t, http.StatusOK, body)

	users, err := c.ListUsers(context.Background())
	if err != nil {
		t.Fatalf("ListUsers() error = %v", err)
	}
	if len(users) != 2 {
		t.Errorf("len(users) = %d, want 2", len(users))
	}
}
