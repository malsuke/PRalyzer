package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/go-github/v77/github"
	"github.com/malsuke/PRalyzer/internal/llm"
)

type ReviewCommentJson struct {
	IssueComments  []llm.PullRequestCommentsPayload `json:"issue_comments"`
	ReviewComments []PullRequestReviewPayload       `json:"review_comments"`
}

type PullRequestReviewPayload struct {
	CommentID int              `json:"id"`
	UserName  string           `json:"user_name"`
	Path      string           `json:"path"`
	DiffHunk  string           `json:"diff_hunk"`
	Body      string           `json:"body"`
	CreatedAt github.Timestamp `json:"created_at"`
	UpdatedAt github.Timestamp `json:"updated_at"`
}

type PRComments struct {
	IssueComments  []*github.IssueComment       `json:"issue_comments"`
	ReviewComments []*github.PullRequestComment `json:"review_comments"`
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: go run main.go <input-directory>")
	}

	inputDir := os.Args[1]
	outputDir := "output"

	// outputディレクトリを作成
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	// 再帰的にJSONファイルを処理
	err := filepath.Walk(inputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// JSONファイルのみを処理
		if !strings.HasSuffix(strings.ToLower(path), ".json") {
			return nil
		}

		// .processed_prs.jsonなどの特殊ファイルをスキップ
		if strings.HasPrefix(filepath.Base(path), ".") {
			return nil
		}

		fmt.Printf("Processing: %s\n", path)

		// JSONファイルを読み込む
		data, err := ioutil.ReadFile(path)
		if err != nil {
			log.Printf("Failed to read file %s: %v", path, err)
			return nil // エラーがあっても続行
		}

		// PRCommentsとしてパース
		var prComments PRComments
		if err := json.Unmarshal(data, &prComments); err != nil {
			log.Printf("Failed to parse JSON file %s: %v", path, err)
			return nil // エラーがあっても続行
		}

		// ReviewCommentJson形式に変換
		reviewCommentJson := convertToReviewCommentJson(prComments)

		// 出力ファイルパスを決定（入力ディレクトリからの相対パスを維持）
		relPath, err := filepath.Rel(inputDir, path)
		if err != nil {
			log.Printf("Failed to get relative path for %s: %v", path, err)
			return nil
		}

		outputPath := filepath.Join(outputDir, relPath)
		outputDirPath := filepath.Dir(outputPath)

		// 出力ディレクトリを作成
		if err := os.MkdirAll(outputDirPath, 0755); err != nil {
			log.Printf("Failed to create output directory %s: %v", outputDirPath, err)
			return nil
		}

		// JSONファイルに書き込む
		outputData, err := json.MarshalIndent(reviewCommentJson, "", "  ")
		if err != nil {
			log.Printf("Failed to marshal JSON for %s: %v", path, err)
			return nil
		}

		if err := ioutil.WriteFile(outputPath, outputData, 0644); err != nil {
			log.Printf("Failed to write file %s: %v", outputPath, err)
			return nil
		}

		fmt.Printf("  -> Saved to: %s\n", outputPath)
		return nil
	})

	if err != nil {
		log.Fatalf("Error walking directory: %v", err)
	}

	fmt.Println("\nDone!")
}

// convertToReviewCommentJson はPRCommentsをReviewCommentJson形式に変換する
func convertToReviewCommentJson(prComments PRComments) ReviewCommentJson {
	// internal/llmパッケージの関数を使って変換
	payloads := llm.ConvertPRCommentsToPayload(prComments.IssueComments, prComments.ReviewComments)

	var issueComments []llm.PullRequestCommentsPayload
	var reviewComments []PullRequestReviewPayload

	// Review Commentsの元データをマップに保存（PathとDiffHunkを取得するため）
	reviewCommentMap := make(map[int]*github.PullRequestComment)
	for _, comment := range prComments.ReviewComments {
		if comment != nil && comment.ID != nil {
			reviewCommentMap[int(*comment.ID)] = comment
		}
	}

	// payloadsをTypeで分類
	for _, payload := range payloads {
		if payload.Type == "issue_comment" {
			issueComments = append(issueComments, payload)
		} else if payload.Type == "review_comment" {
			// Review Commentの場合は、元のデータからPathとDiffHunkを取得
			originalComment := reviewCommentMap[payload.CommentID]
			path := ""
			diffHunk := ""
			if originalComment != nil {
				if originalComment.Path != nil {
					path = *originalComment.Path
				}
				if originalComment.DiffHunk != nil {
					diffHunk = *originalComment.DiffHunk
				}
			}

			reviewComments = append(reviewComments, PullRequestReviewPayload{
				CommentID: payload.CommentID,
				UserName:  payload.UserName,
				Path:      path,
				DiffHunk:  diffHunk,
				Body:      payload.Body,
				CreatedAt: payload.CreatedAt,
				UpdatedAt: payload.UpdatedAt,
			})
		}
	}

	return ReviewCommentJson{
		IssueComments:  issueComments,
		ReviewComments: reviewComments,
	}
}

func parsePRCommentsFromJson(str string) (*PRComments, error) {
	var prComments PRComments
	err := json.Unmarshal([]byte(str), &prComments)
	if err != nil {
		return nil, err
	}
	return &prComments, nil
}
