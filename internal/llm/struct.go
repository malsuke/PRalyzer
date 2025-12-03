package llm

import (
	"sort"

	"github.com/google/go-github/v77/github"
)

type PullRequestCommentsPayload struct {
	CommentID int              `json:"id"`
	UserName  string           `json:"user_name"`
	Body      string           `json:"body"`
	Type      string           `json:"type"`
	CreatedAt github.Timestamp `json:"created_at"`
	UpdatedAt github.Timestamp `json:"updated_at"`
}

/*
*[]*IssueCommentをPullRequestCommentsPayloadの配列に変換する
 */
func ConvertToPullRequestCommentsPayload(comments []*github.IssueComment) []PullRequestCommentsPayload {
	payloads := make([]PullRequestCommentsPayload, len(comments))
	for i, comment := range comments {
		var createdAt, updatedAt github.Timestamp
		if comment.CreatedAt != nil {
			createdAt = *comment.CreatedAt
		}
		if comment.UpdatedAt != nil {
			updatedAt = *comment.UpdatedAt
		}
		payloads[i] = PullRequestCommentsPayload{
			CommentID: int(*comment.ID),
			UserName:  *comment.User.Login,
			Body:      *comment.Body,
			Type:      *comment.User.Type,
			CreatedAt: createdAt,
			UpdatedAt: updatedAt,
		}
	}
	return payloads
}

/*
* Issue CommentsとReview Commentsを統合してPullRequestCommentsPayloadの配列に変換する
* 時系列順にソートされた結果を返す
 */
func ConvertPRCommentsToPayload(issueComments []*github.IssueComment, reviewComments []*github.PullRequestComment) []PullRequestCommentsPayload {
	var payloads []PullRequestCommentsPayload

	// Issue Commentsを変換
	for _, comment := range issueComments {
		if comment == nil {
			continue
		}

		var createdAt, updatedAt github.Timestamp
		if comment.CreatedAt != nil {
			createdAt = *comment.CreatedAt
		}
		if comment.UpdatedAt != nil {
			updatedAt = *comment.UpdatedAt
		}

		userName := ""
		if comment.User != nil && comment.User.Login != nil {
			userName = *comment.User.Login
		}

		body := ""
		if comment.Body != nil {
			body = *comment.Body
		}

		commentID := 0
		if comment.ID != nil {
			commentID = int(*comment.ID)
		}

		payloads = append(payloads, PullRequestCommentsPayload{
			CommentID: commentID,
			UserName:  userName,
			Body:      body,
			Type:      "issue_comment",
			CreatedAt: createdAt,
			UpdatedAt: updatedAt,
		})
	}

	// Review Commentsを変換
	for _, comment := range reviewComments {
		if comment == nil {
			continue
		}

		var createdAt, updatedAt github.Timestamp
		if comment.CreatedAt != nil {
			createdAt = *comment.CreatedAt
		}
		if comment.UpdatedAt != nil {
			updatedAt = *comment.UpdatedAt
		}

		userName := ""
		if comment.User != nil && comment.User.Login != nil {
			userName = *comment.User.Login
		}

		body := ""
		if comment.Body != nil {
			body = *comment.Body
		}

		commentID := 0
		if comment.ID != nil {
			commentID = int(*comment.ID)
		}

		payloads = append(payloads, PullRequestCommentsPayload{
			CommentID: commentID,
			UserName:  userName,
			Body:      body,
			Type:      "review_comment",
			CreatedAt: createdAt,
			UpdatedAt: updatedAt,
		})
	}

	// 時系列順にソート
	sortPayloadsByTime(payloads)

	return payloads
}

// sortPayloadsByTime はPullRequestCommentsPayloadを時系列順にソートする
func sortPayloadsByTime(payloads []PullRequestCommentsPayload) {
	sort.Slice(payloads, func(i, j int) bool {
		return payloads[i].CreatedAt.Time.Before(payloads[j].CreatedAt.Time)
	})
}
