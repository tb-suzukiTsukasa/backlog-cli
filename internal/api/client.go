package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// httpDoer は HTTP リクエストを実行するインターフェース（テスト用モック注入のため）。
type httpDoer interface {
	Do(*http.Request) (*http.Response, error)
}

// Client は Backlog API クライアント。
type Client struct {
	host       string
	apiKey     string
	httpClient httpDoer
}

// NewClient は Backlog API クライアントを生成する。
// host は "myteam.backlog.com" 形式。
func NewClient(host, apiKey string) (*Client, error) {
	if host == "" {
		return nil, fmt.Errorf("ホスト名が指定されていません")
	}
	if apiKey == "" {
		return nil, fmt.Errorf("APIキーが指定されていません")
	}
	return &Client{host: host, apiKey: apiKey, httpClient: &http.Client{Timeout: 30 * time.Second}}, nil
}

func (c *Client) baseURL() string {
	return "https://" + c.host + "/api/v2"
}

// get は GET リクエストを発行し、レスポンス JSON を out にデコードする。
func (c *Client) get(ctx context.Context, path string, params url.Values, out interface{}) error {
	if params == nil {
		params = url.Values{}
	}
	params.Set("apiKey", c.apiKey)
	u := c.baseURL() + path + "?" + params.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return fmt.Errorf("リクエスト作成エラー: %w", err)
	}
	return c.do(req, out)
}

// post は POST リクエストを発行し、レスポンス JSON を out にデコードする。
func (c *Client) post(ctx context.Context, path string, form url.Values, out interface{}) error {
	u := c.baseURL() + path + "?apiKey=" + url.QueryEscape(c.apiKey)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, strings.NewReader(form.Encode()))
	if err != nil {
		return fmt.Errorf("リクエスト作成エラー: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return c.do(req, out)
}

func (c *Client) do(req *http.Request, out interface{}) error {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("API リクエストエラー: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return convertError(resp.StatusCode, body)
	}

	if out != nil {
		if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
			return fmt.Errorf("レスポンス解析エラー: %w", err)
		}
	} else {
		_, _ = io.Copy(io.Discard, resp.Body)
	}
	return nil
}

// convertError は HTTP ステータスコードをユーザーフレンドリーなエラーに変換する。
// Backlog API のエラー JSON ({errors:[{message:...}]}) があれば詳細を付加する。
func convertError(statusCode int, body []byte) error {
	var berr struct {
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}
	detail := ""
	if json.Unmarshal(body, &berr) == nil && len(berr.Errors) > 0 && berr.Errors[0].Message != "" {
		detail = " (" + berr.Errors[0].Message + ")"
	}

	switch statusCode {
	case 401:
		return fmt.Errorf("認証エラー: APIキーが無効です。`bk auth login` で再設定してください%s", detail)
	case 403:
		return fmt.Errorf("権限エラー: このリソースへのアクセス権限がありません%s", detail)
	case 404:
		return fmt.Errorf("リソースが見つかりません%s", detail)
	case 429:
		return fmt.Errorf("レート制限に達しました。しばらく待ってから再試行してください%s", detail)
	default:
		if detail != "" {
			return fmt.Errorf("Backlog API エラー (status %d)%s", statusCode, detail)
		}
		return fmt.Errorf("Backlog API エラー (status %d): %s", statusCode, string(body))
	}
}

// ListUsers はスペースの全ユーザー一覧を返す。
func (c *Client) ListUsers(ctx context.Context) ([]*User, error) {
	var users []*User
	if err := c.get(ctx, "/users", nil, &users); err != nil {
		return nil, err
	}
	return users, nil
}

// GetRepository はリポジトリ情報を返す。
func (c *Client) GetRepository(ctx context.Context, projectKey, repoName string) (*GitRepository, error) {
	path := fmt.Sprintf("/projects/%s/git/repositories/%s", projectKey, repoName)
	var repo GitRepository
	if err := c.get(ctx, path, nil, &repo); err != nil {
		return nil, err
	}
	return &repo, nil
}
