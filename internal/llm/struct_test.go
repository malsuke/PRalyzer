package llm

import (
	"testing"
	"time"

	"github.com/google/go-github/v77/github"
	"github.com/stretchr/testify/assert"
)

func TestConvertToPullRequestCommentsPayload(t *testing.T) {
	tests := []struct {
		name     string
		comments []*github.IssueComment
		want     []PullRequestCommentsPayload
	}{
		{
			name:     "空のコメント配列",
			comments: []*github.IssueComment{},
			want:     []PullRequestCommentsPayload{},
		},
		{
			name: "単一のコメント",
			comments: []*github.IssueComment{
				{
					ID:        int64Ptr(12345),
					User:      &github.User{Login: stringPtr("testuser"), Type: stringPtr("User")},
					Body:      stringPtr("これはテストコメントです"),
					CreatedAt: timestampPtr(time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)),
					UpdatedAt: timestampPtr(time.Date(2024, 1, 2, 12, 0, 0, 0, time.UTC)),
				},
			},
			want: []PullRequestCommentsPayload{
				{
					CommentID: 12345,
					UserName:  "testuser",
					Body:      "これはテストコメントです",
					Type:      "User",
					CreatedAt: github.Timestamp{Time: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)},
					UpdatedAt: github.Timestamp{Time: time.Date(2024, 1, 2, 12, 0, 0, 0, time.UTC)},
				},
			},
		},
		{
			name: "複数のコメント",
			comments: []*github.IssueComment{
				{
					ID:        int64Ptr(12345),
					User:      &github.User{Login: stringPtr("user1"), Type: stringPtr("User")},
					Body:      stringPtr("最初のコメント"),
					CreatedAt: timestampPtr(time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)),
					UpdatedAt: timestampPtr(time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)),
				},
				{
					ID:        int64Ptr(67890),
					User:      &github.User{Login: stringPtr("user2"), Type: stringPtr("User")},
					Body:      stringPtr("2番目のコメント"),
					CreatedAt: timestampPtr(time.Date(2024, 1, 2, 12, 0, 0, 0, time.UTC)),
					UpdatedAt: timestampPtr(time.Date(2024, 1, 2, 12, 0, 0, 0, time.UTC)),
				},
				{
					ID:        int64Ptr(11111),
					User:      &github.User{Login: stringPtr("user3"), Type: stringPtr("User")},
					Body:      stringPtr("3番目のコメント"),
					CreatedAt: timestampPtr(time.Date(2024, 1, 3, 12, 0, 0, 0, time.UTC)),
					UpdatedAt: timestampPtr(time.Date(2024, 1, 3, 12, 0, 0, 0, time.UTC)),
				},
			},
			want: []PullRequestCommentsPayload{
				{
					CommentID: 12345,
					UserName:  "user1",
					Body:      "最初のコメント",
					Type:      "User",
					CreatedAt: github.Timestamp{Time: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)},
					UpdatedAt: github.Timestamp{Time: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)},
				},
				{
					CommentID: 67890,
					UserName:  "user2",
					Body:      "2番目のコメント",
					Type:      "User",
					CreatedAt: github.Timestamp{Time: time.Date(2024, 1, 2, 12, 0, 0, 0, time.UTC)},
					UpdatedAt: github.Timestamp{Time: time.Date(2024, 1, 2, 12, 0, 0, 0, time.UTC)},
				},
				{
					CommentID: 11111,
					UserName:  "user3",
					Body:      "3番目のコメント",
					Type:      "User",
					CreatedAt: github.Timestamp{Time: time.Date(2024, 1, 3, 12, 0, 0, 0, time.UTC)},
					UpdatedAt: github.Timestamp{Time: time.Date(2024, 1, 3, 12, 0, 0, 0, time.UTC)},
				},
			},
		},
		{
			name: "空のボディを持つコメント",
			comments: []*github.IssueComment{
				{
					ID:        int64Ptr(99999),
					User:      &github.User{Login: stringPtr("emptyuser"), Type: stringPtr("User")},
					Body:      stringPtr(""),
					CreatedAt: timestampPtr(time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)),
					UpdatedAt: timestampPtr(time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)),
				},
			},
			want: []PullRequestCommentsPayload{
				{
					CommentID: 99999,
					UserName:  "emptyuser",
					Body:      "",
					Type:      "User",
					CreatedAt: github.Timestamp{Time: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)},
					UpdatedAt: github.Timestamp{Time: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ConvertToPullRequestCommentsPayload(tt.comments)
			assert.Equal(t, tt.want, got)
		})
	}
}

// ヘルパー関数: int64のポインタを作成
func int64Ptr(v int64) *int64 {
	return &v
}

// ヘルパー関数: stringのポインタを作成
func stringPtr(v string) *string {
	return &v
}

// ヘルパー関数: github.Timestampのポインタを作成
func timestampPtr(t time.Time) *github.Timestamp {
	return &github.Timestamp{Time: t}
}
