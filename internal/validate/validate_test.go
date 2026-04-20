package validate

import (
	"testing"
)

func TestProjectName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid simple", "myproject", false},
		{"valid with hyphen", "my-project", false},
		{"valid with underscore", "my_project", false},
		{"valid with numbers", "project123", false},
		{"valid single char", "a", false},
		{"empty", "", true},
		{"starts with hyphen", "-bad", true},
		{"starts with underscore", "_bad", true},
		{"contains spaces", "my project", true},
		{"contains dots", "my.project", true},
		{"path traversal", "../../etc", true},
		{"slash", "a/b", true},
		{"backslash", "a\\b", true},
		{"too long", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", true}, // 67 chars
		{"max length", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", false},  // 64 chars
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ProjectName(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ProjectName(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestGitURL(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"empty is ok", "", false},
		{"valid https", "https://github.com/user/repo.git", false},
		{"valid git protocol", "git://github.com/user/repo.git", false},
		{"valid file protocol", "file:///C:/path/to/repo", false},
		{"valid windows path", "C:/path/to/repo", false},
		{"valid unix path", "/path/to/repo", false},
		{"http not allowed", "http://github.com/user/repo.git", true},
		{"ssh not allowed", "ssh://git@github.com/user/repo.git", true},
		{"file is allowed", "file:///etc/passwd", false},
		{"no host on remote not allowed", "https://", true},
		{"embedded credentials", "https://user:pass@github.com/repo.git", true},
		{"garbage with spaces not allowed", "not a url at all ;;; !!!", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := GitURL(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("GitURL(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestWorkspacePath(t *testing.T) {
	tests := []struct {
		name        string
		base        string
		project     string
		wantErr     bool
	}{
		{"valid", "./workspaces", "myproject", false},
		{"traversal dots", "./workspaces", "../../etc", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := WorkspacePath(tt.base, tt.project)
			if (err != nil) != tt.wantErr {
				t.Errorf("WorkspacePath(%q, %q) error = %v, wantErr %v", tt.base, tt.project, err, tt.wantErr)
			}
		})
	}
}

func TestContainerID(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid hex id", "abc123def456", false},
		{"valid name", "my-container", false},
		{"valid with dots", "my.container", false},
		{"empty", "", true},
		{"starts with special", "-bad", true},
		{"has spaces", "my container", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ContainerID(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ContainerID(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}
