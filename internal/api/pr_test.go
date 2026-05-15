package api

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"
)

const prListJSON = `[
  {"id":1,"projectId":10,"repositoryId":20,"number":1,"summary":"Fix bug","base":"main","branch":"fix/bug","status":{"id":1,"name":"Open"},"createdUser":{"id":1,"name":"Alice"}},
  {"id":2,"projectId":10,"repositoryId":20,"number":2,"summary":"Add feature","base":"main","branch":"feat/new","status":{"id":1,"name":"Open"},"createdUser":{"id":2,"name":"Bob"}}
]`

const prSingleJSON = `{"id":1,"projectId":10,"repositoryId":20,"number":1,"summary":"Fix bug","description":"Details","base":"main","branch":"fix/bug","status":{"id":1,"name":"Open"},"createdUser":{"id":1,"name":"Alice"}}`

const prCreatedJSON = `{"id":3,"projectId":10,"repositoryId":20,"number":3,"summary":"New PR","description":"","base":"main","branch":"feature/x","status":{"id":1,"name":"Open"},"createdUser":{"id":1,"name":"Alice"}}`

func TestListPullRequests(t *testing.T) {
	c := newMockClient(t, http.StatusOK, prListJSON)

	prs, err := c.ListPullRequests(context.Background(), "PROJ", "repo", 10, PRListOptions{})
	if err != nil {
		t.Fatalf("ListPullRequests() error = %v", err)
	}
	if len(prs) != 2 {
		t.Errorf("len(prs) = %d, want 2", len(prs))
	}
	if prs[0].Summary != "Fix bug" {
		t.Errorf("prs[0].Summary = %q, want %q", prs[0].Summary, "Fix bug")
	}
}

func TestGetPullRequest(t *testing.T) {
	c := newMockClient(t, http.StatusOK, prSingleJSON)

	pr, err := c.GetPullRequest(context.Background(), "PROJ", "repo", 1)
	if err != nil {
		t.Fatalf("GetPullRequest() error = %v", err)
	}
	if pr.Number != 1 {
		t.Errorf("pr.Number = %d, want 1", pr.Number)
	}
	if pr.Summary != "Fix bug" {
		t.Errorf("pr.Summary = %q, want %q", pr.Summary, "Fix bug")
	}
}

func TestCreatePullRequest(t *testing.T) {
	c := newMockClient(t, http.StatusCreated, prCreatedJSON)

	pr, err := c.CreatePullRequest(context.Background(), "PROJ", "repo", "New PR", "", "main", "feature/x", PRCreateOptions{})
	if err != nil {
		t.Fatalf("CreatePullRequest() error = %v", err)
	}
	if pr.Summary != "New PR" {
		t.Errorf("pr.Summary = %q, want %q", pr.Summary, "New PR")
	}
}

func TestCreatePullRequest_MissingTitle(t *testing.T) {
	c := newMockClient(t, http.StatusCreated, prCreatedJSON)

	_, err := c.CreatePullRequest(context.Background(), "PROJ", "repo", "", "", "main", "feature/x", PRCreateOptions{})
	if err == nil {
		t.Fatal("expected error for missing title, got nil")
	}
}

func TestCreatePullRequest_MissingBase(t *testing.T) {
	c := newMockClient(t, http.StatusCreated, prCreatedJSON)

	_, err := c.CreatePullRequest(context.Background(), "PROJ", "repo", "Title", "", "", "feature/x", PRCreateOptions{})
	if err == nil {
		t.Fatal("expected error for missing base, got nil")
	}
}

func TestCreatePullRequest_EmptyBranch(t *testing.T) {
	c := newMockClient(t, http.StatusCreated, prCreatedJSON)
	_, err := c.CreatePullRequest(context.Background(), "PROJ", "repo", "Title", "", "main", "", PRCreateOptions{})
	if err == nil {
		t.Fatal("expected error for empty branch, got nil")
	}
}

func TestListPullRequests_Pagination(t *testing.T) {
	// 100件フルページ + 5件の2ページ構成でlimit=105を検証
	firstPage := generatePRListJSON(1, 100)
	secondPage := generatePRListJSON(101, 5)

	mock := &multiMockHTTPClient{
		responses: []struct {
			statusCode int
			body       string
		}{
			{http.StatusOK, firstPage},
			{http.StatusOK, secondPage},
		},
	}
	c := &Client{host: "example.backlog.com", apiKey: "test-key", httpClient: mock}

	prs, err := c.ListPullRequests(context.Background(), "PROJ", "repo", 105, PRListOptions{})
	if err != nil {
		t.Fatalf("ListPullRequests() error = %v", err)
	}
	if len(prs) != 105 {
		t.Errorf("len(prs) = %d, want 105", len(prs))
	}
	if prs[0].Number != 1 {
		t.Errorf("prs[0].Number = %d, want 1", prs[0].Number)
	}
	if prs[104].Number != 105 {
		t.Errorf("prs[104].Number = %d, want 105", prs[104].Number)
	}
}

// generatePRListJSON は start から count 件分の PR JSON 配列を生成する。
func generatePRListJSON(start, count int) string {
	var sb strings.Builder
	sb.WriteString("[")
	for i := 0; i < count; i++ {
		if i > 0 {
			sb.WriteString(",")
		}
		n := start + i
		fmt.Fprintf(&sb, `{"id":%d,"number":%d,"summary":"PR %d","base":"main","branch":"branch-%d","status":{"id":1,"name":"Open"}}`, n, n, n, n)
	}
	sb.WriteString("]")
	return sb.String()
}

func TestAddPullRequestComment(t *testing.T) {
	commentJSON := `{"id":1,"content":"LGTM","createdUser":{"id":1,"name":"Alice"}}`
	c := newMockClient(t, http.StatusCreated, commentJSON)

	err := c.AddPullRequestComment(context.Background(), "PROJ", "repo", 1, "LGTM")
	if err != nil {
		t.Fatalf("AddPullRequestComment() error = %v", err)
	}
}
