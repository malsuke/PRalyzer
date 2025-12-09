package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/malsuke/PRalyzer/internal/llm"
)

type ReviewCommentJson struct {
	IssueComments  []llm.PullRequestCommentsPayload `json:"issue_comments"`
	ReviewComments []ReviewComment                  `json:"review_comments"`
}

type ReviewComment struct {
	CommentID int    `json:"id"`
	UserName  string `json:"user_name"`
	Path      string `json:"path"`
	DiffHunk  string `json:"diff_hunk"`
	Body      string `json:"body"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func main() {
	outputDir := "data"

	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		log.Fatalf("Output directory does not exist: %s", outputDir)
	}

	// 再帰的にJSONファイルを処理
	err := filepath.Walk(outputDir, func(path string, info os.FileInfo, err error) error {
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

		// ReviewCommentJsonとしてパース
		var reviewCommentJson ReviewCommentJson
		if err := json.Unmarshal(data, &reviewCommentJson); err != nil {
			log.Printf("Failed to parse JSON file %s: %v", path, err)
			return nil // エラーがあっても続行
		}

		// issue_commentsから「## Stats from current PR」で始まるコメントを削除
		var filteredComments []llm.PullRequestCommentsPayload
		removedCount := 0
		for _, comment := range reviewCommentJson.IssueComments {
			// bodyが「## Stats from current PR」で始まるかチェック
			bodyTrimmed := strings.TrimSpace(comment.Body)
			if strings.HasPrefix(bodyTrimmed, "## Stats from current PR") {
				removedCount++
				continue
			}
			filteredComments = append(filteredComments, comment)
		}

		if removedCount > 0 {
			reviewCommentJson.IssueComments = filteredComments

			// JSONファイルに書き込む
			outputData, err := json.MarshalIndent(reviewCommentJson, "", "  ")
			if err != nil {
				log.Printf("Failed to marshal JSON for %s: %v", path, err)
				return nil
			}

			if err := ioutil.WriteFile(path, outputData, 0644); err != nil {
				log.Printf("Failed to write file %s: %v", path, err)
				return nil
			}

			fmt.Printf("  -> Removed %d comment(s) from %s\n", removedCount, path)
		} else {
			fmt.Printf("  -> No comments to remove in %s\n", path)
		}

		return nil
	})

	if err != nil {
		log.Fatalf("Error walking directory: %v", err)
	}

	fmt.Println("\nDone!")
}
