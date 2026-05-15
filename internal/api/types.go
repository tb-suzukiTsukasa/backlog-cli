package api

// User は Backlog ユーザーを表す。
type User struct {
	ID     int    `json:"id"`
	UserID string `json:"userId"`
	Name   string `json:"name"`
}

// PRStatus は PR のステータスを表す。
type PRStatus struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// PullRequest は Backlog の PR を表す。
type PullRequest struct {
	ID           int       `json:"id"`
	ProjectID    int       `json:"projectId"`
	RepositoryID int       `json:"repositoryId"`
	Number       int       `json:"number"`
	Summary      string    `json:"summary"`
	Description  string    `json:"description"`
	Base         string    `json:"base"`
	Branch       string    `json:"branch"`
	BaseCommit   string    `json:"baseCommit"`
	BranchCommit string    `json:"branchCommit"`
	MergeCommit  string    `json:"mergeCommit"`
	Status       *PRStatus `json:"status"`
	CreatedUser  *User     `json:"createdUser"`
}

// GitRepository は Backlog の Git リポジトリを表す。
type GitRepository struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// Project は Backlog のプロジェクトを表す。
type Project struct {
	ID         int    `json:"id"`
	ProjectKey string `json:"projectKey"`
	Name       string `json:"name"`
	UseGit     bool   `json:"useGit"`
}

// IssueType は課題の種別。
type IssueType struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// Priority は優先度。
type Priority struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// IssueStatus は課題のステータス。
type IssueStatus struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// Issue は Backlog の課題を表す。
type Issue struct {
	ID          int          `json:"id"`
	ProjectID   int          `json:"projectId"`
	IssueKey    string       `json:"issueKey"`
	IssueType   *IssueType   `json:"issueType"`
	Summary     string       `json:"summary"`
	Description string       `json:"description"`
	Status      *IssueStatus `json:"status"`
	Priority    *Priority    `json:"priority"`
	Assignee    *User        `json:"assignee"`
	CreatedUser *User        `json:"createdUser"`
}
