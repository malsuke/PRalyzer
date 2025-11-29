package github

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseRepository(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		owner   string
		repo    string
		wantErr bool
	}{
		{
			name:  "owner/name",
			input: "owner/repo",
			owner: "owner",
			repo:  "repo",
		},
		{
			name:  "https url",
			input: "https://github.com/owner/repo",
			owner: "owner",
			repo:  "repo",
		},
		{
			name:  "git ssh url",
			input: "git@github.com:owner/repo.git",
			owner: "owner",
			repo:  "repo",
		},
		{
			name:  "PR URL",
			input: "https://github.com/owner/repo/pull/123",
			owner: "owner",
			repo:  "repo",
		},
		{
			name:  "Issue URL",
			input: "https://github.com/owner/repo/issues/456",
			owner: "owner",
			repo:  "repo",
		},
		{
			name:    "invalid",
			input:   "owner-only",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			owner, repo, err := ParseRepository(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.owner, owner)
			assert.Equal(t, tt.repo, repo)
		})
	}
}
