package api

import (
	"context"
	"fmt"
)

// ListProjects は自分が参加している全プロジェクト一覧を返す。
func (c *Client) ListProjects(ctx context.Context) ([]*Project, error) {
	var projects []*Project
	if err := c.get(ctx, "/projects", nil, &projects); err != nil {
		return nil, fmt.Errorf("プロジェクト一覧の取得に失敗しました: %w", err)
	}
	return projects, nil
}

// ListRepositories はプロジェクトのリポジトリ一覧を返す。
func (c *Client) ListRepositories(ctx context.Context, projectKey string) ([]*GitRepository, error) {
	path := fmt.Sprintf("/projects/%s/git/repositories", projectKey)
	var repos []*GitRepository
	if err := c.get(ctx, path, nil, &repos); err != nil {
		return nil, fmt.Errorf("リポジトリ一覧の取得に失敗しました: %w", err)
	}
	return repos, nil
}
