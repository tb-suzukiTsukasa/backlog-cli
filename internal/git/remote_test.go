package git

import "testing"

func TestParseBacklogURL_HTTPS(t *testing.T) {
	tests := []struct {
		name        string
		url         string
		wantSpace   string
		wantHost    string
		wantProject string
		wantRepo    string
	}{
		{
			name:        "backlog.com with .git",
			url:         "https://myteam.backlog.com/git/PROJ/myrepo.git",
			wantSpace:   "myteam",
			wantHost:    "myteam.backlog.com",
			wantProject: "PROJ",
			wantRepo:    "myrepo",
		},
		{
			name:        "backlog.jp without .git",
			url:         "https://myteam.backlog.jp/git/MYPROJ/repo",
			wantSpace:   "myteam",
			wantHost:    "myteam.backlog.jp",
			wantProject: "MYPROJ",
			wantRepo:    "repo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := parseBacklogURL(tt.url)
			if err != nil {
				t.Fatalf("parseBacklogURL() error = %v", err)
			}
			if r.SpaceName != tt.wantSpace {
				t.Errorf("SpaceName = %q, want %q", r.SpaceName, tt.wantSpace)
			}
			if r.Host != tt.wantHost {
				t.Errorf("Host = %q, want %q", r.Host, tt.wantHost)
			}
			if r.Project != tt.wantProject {
				t.Errorf("Project = %q, want %q", r.Project, tt.wantProject)
			}
			if r.Repo != tt.wantRepo {
				t.Errorf("Repo = %q, want %q", r.Repo, tt.wantRepo)
			}
		})
	}
}

func TestParseBacklogURL_SSH(t *testing.T) {
	tests := []struct {
		name        string
		url         string
		wantSpace   string
		wantHost    string
		wantProject string
		wantRepo    string
	}{
		{
			name:        "SSH backlog.com with .git",
			url:         "git@myteam.git.backlog.com:PROJ/myrepo.git",
			wantSpace:   "myteam",
			wantHost:    "myteam.backlog.com",
			wantProject: "PROJ",
			wantRepo:    "myrepo",
		},
		{
			name:        "SSH backlog.jp without .git",
			url:         "git@myteam.git.backlog.jp:MYPROJ/repo",
			wantSpace:   "myteam",
			wantHost:    "myteam.backlog.jp",
			wantProject: "MYPROJ",
			wantRepo:    "repo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := parseBacklogURL(tt.url)
			if err != nil {
				t.Fatalf("parseBacklogURL() error = %v", err)
			}
			if r.SpaceName != tt.wantSpace {
				t.Errorf("SpaceName = %q, want %q", r.SpaceName, tt.wantSpace)
			}
			if r.Host != tt.wantHost {
				t.Errorf("Host = %q, want %q", r.Host, tt.wantHost)
			}
			if r.Project != tt.wantProject {
				t.Errorf("Project = %q, want %q", r.Project, tt.wantProject)
			}
			if r.Repo != tt.wantRepo {
				t.Errorf("Repo = %q, want %q", r.Repo, tt.wantRepo)
			}
		})
	}
}

func TestParseBacklogURL_NonBacklog(t *testing.T) {
	urls := []string{
		"https://github.com/user/repo.git",
		"git@github.com:user/repo.git",
		"https://gitlab.com/user/repo.git",
	}
	for _, url := range urls {
		t.Run(url, func(t *testing.T) {
			_, err := parseBacklogURL(url)
			if err == nil {
				t.Fatalf("expected error for non-Backlog URL %q, got nil", url)
			}
		})
	}
}
