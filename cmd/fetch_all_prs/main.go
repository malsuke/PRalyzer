package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/malsuke/PRalyzer/internal/github"
)

func main() {
	if len(os.Args) < 3 {
		log.Fatal("Usage: go run cmd/fetch_all_prs/main.go <repository-url> <github-pat>")
	}

	repoURL := os.Args[1]
	githubPAT := os.Args[2]

	client, err := github.NewClient(githubPAT, repoURL, nil)
	if err != nil {
		log.Fatalf("Failed to create GitHub client: %v", err)
	}

	fmt.Printf("Fetching all pull requests from %s/%s...\n", client.Owner, client.Name)

	// すべてのPRを取得
	prs, err := client.ListAllPullRequests()
	if err != nil {
		log.Fatalf("Failed to list pull requests: %v", err)
	}

	fmt.Printf("Found %d pull requests\n", len(prs))

	// 出力ディレクトリを作成
	outputDir := filepath.Join("data", client.Owner, client.Name)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	// 各PRをJSONファイルに保存
	savedCount := 0
	for _, pr := range prs {
		if pr.Number == nil {
			log.Printf("Skipping PR with nil number")
			continue
		}

		prNumber := *pr.Number
		outputPath := filepath.Join(outputDir, fmt.Sprintf("%d.json", prNumber))

		// PRをJSONにエンコード
		data, err := json.MarshalIndent(pr, "", "  ")
		if err != nil {
			log.Printf("Failed to marshal PR #%d: %v", prNumber, err)
			continue
		}

		// ファイルに書き込む
		if err := os.WriteFile(outputPath, data, 0644); err != nil {
			log.Printf("Failed to write PR #%d to file: %v", prNumber, err)
			continue
		}

		savedCount++
		fmt.Printf("Saved PR #%d to %s\n", prNumber, outputPath)
	}

	fmt.Printf("\nDone! Saved %d pull requests to %s\n", savedCount, outputDir)
}

