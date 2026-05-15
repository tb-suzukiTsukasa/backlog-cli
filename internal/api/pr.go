package api

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
)

// PRListOptions は PR一覧の取得オプション。
type PRListOptions struct {
	StatusIDs []int
	Count     int
	Offset    int
}

// PRCreateOptions は PR 作成オプション。
type PRCreateOptions struct {
	// NotifiedUserIDs はレビュアーとして通知するユーザーID一覧。
	NotifiedUserIDs []int
}

// ListPullRequests は PR一覧を返す。limit が 100 超の場合は自動ページネーション。
func (c *Client) ListPullRequests(ctx context.Context, projectKey, repoName string, limit int, opts PRListOptions) ([]*PullRequest, error) {
	path := fmt.Sprintf("/projects/%s/git/repositories/%s/pullRequests", projectKey, repoName)
	var all []*PullRequest
	pageSize := 100
	offset := opts.Offset

	for {
		remaining := limit - len(all)
		if remaining <= 0 {
			break
		}
		count := pageSize
		if remaining < pageSize {
			count = remaining
		}

		params := url.Values{}
		params.Set("count", strconv.Itoa(count))
		params.Set("offset", strconv.Itoa(offset))
		for _, id := range opts.StatusIDs {
			params.Add("statusId[]", strconv.Itoa(id))
		}

		var prs []*PullRequest
		if err := c.get(ctx, path, params, &prs); err != nil {
			return nil, err
		}
		all = append(all, prs...)

		if len(prs) < count {
			break
		}
		offset += count
	}
	return all, nil
}

// GetPullRequest は指定番号の PR 詳細を返す。
func (c *Client) GetPullRequest(ctx context.Context, projectKey, repoName string, number int) (*PullRequest, error) {
	path := fmt.Sprintf("/projects/%s/git/repositories/%s/pullRequests/%d", projectKey, repoName, number)
	var pr PullRequest
	if err := c.get(ctx, path, nil, &pr); err != nil {
		return nil, err
	}
	return &pr, nil
}

// CreatePullRequest は PR を作成する。
func (c *Client) CreatePullRequest(ctx context.Context, projectKey, repoName, title, description, base, branch string, opts PRCreateOptions) (*PullRequest, error) {
	if title == "" {
		return nil, fmt.Errorf("PR タイトルを指定してください")
	}
	if base == "" {
		return nil, fmt.Errorf("マージ先ブランチ (--base) を指定してください")
	}
	if branch == "" {
		return nil, fmt.Errorf("ソースブランチを指定してください")
	}

	path := fmt.Sprintf("/projects/%s/git/repositories/%s/pullRequests", projectKey, repoName)
	form := url.Values{}
	form.Set("summary", title)
	form.Set("description", description)
	form.Set("base", base)
	form.Set("branch", branch)
	for _, id := range opts.NotifiedUserIDs {
		form.Add("notifiedUserId[]", strconv.Itoa(id))
	}

	var pr PullRequest
	if err := c.post(ctx, path, form, &pr); err != nil {
		return nil, err
	}
	return &pr, nil
}

// AddPullRequestComment は PR にコメントを追加する。
func (c *Client) AddPullRequestComment(ctx context.Context, projectKey, repoName string, number int, content string) error {
	path := fmt.Sprintf("/projects/%s/git/repositories/%s/pullRequests/%d/comments", projectKey, repoName, number)
	form := url.Values{}
	form.Set("content", content)
	return c.post(ctx, path, form, nil)
}
