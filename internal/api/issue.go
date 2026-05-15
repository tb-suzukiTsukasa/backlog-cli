package api

import (
	"context"
	"net/url"
	"strconv"
)

// IssueListOptions は課題一覧の取得オプション。
type IssueListOptions struct {
	StatusIDs []int
	Offset    int
}

// GetProject はプロジェクトキーからプロジェクト情報を返す。
func (c *Client) GetProject(ctx context.Context, projectKey string) (*Project, error) {
	var project Project
	if err := c.get(ctx, "/projects/"+projectKey, nil, &project); err != nil {
		return nil, err
	}
	return &project, nil
}

// ListIssues は課題一覧を返す。limit が 100 超の場合は自動ページネーション。
func (c *Client) ListIssues(ctx context.Context, projectID, limit int, opts IssueListOptions) ([]*Issue, error) {
	var all []*Issue
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
		params.Add("projectId[]", strconv.Itoa(projectID))
		params.Set("count", strconv.Itoa(count))
		params.Set("offset", strconv.Itoa(offset))
		for _, id := range opts.StatusIDs {
			params.Add("statusId[]", strconv.Itoa(id))
		}

		var issues []*Issue
		if err := c.get(ctx, "/issues", params, &issues); err != nil {
			return nil, err
		}
		all = append(all, issues...)

		if len(issues) < count {
			break
		}
		offset += count
	}
	return all, nil
}

// GetIssue は課題キーまたは ID から課題詳細を返す。
func (c *Client) GetIssue(ctx context.Context, issueIDOrKey string) (*Issue, error) {
	var issue Issue
	if err := c.get(ctx, "/issues/"+issueIDOrKey, nil, &issue); err != nil {
		return nil, err
	}
	return &issue, nil
}

// AddIssueComment は課題にコメントを追加する。
func (c *Client) AddIssueComment(ctx context.Context, issueIDOrKey, content string) error {
	form := url.Values{}
	form.Set("content", content)
	return c.post(ctx, "/issues/"+issueIDOrKey+"/comments", form, nil)
}
