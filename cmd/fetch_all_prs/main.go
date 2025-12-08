package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

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

	fmt.Printf("========================================\n")
	fmt.Printf("Fetching all pull requests from %s/%s\n", client.Owner, client.Name)
	fmt.Printf("========================================\n\n")

	startTime := time.Now()

	// すべてのPRを取得
	prs, err := client.ListAllPullRequests()
	if err != nil {
		log.Fatalf("Failed to list pull requests: %v", err)
	}

	fetchDuration := time.Since(startTime)
	fmt.Printf("\n✓ Successfully fetched %d pull requests in %v\n", len(prs), fetchDuration.Round(time.Second))
	fmt.Printf("\n========================================\n")
	fmt.Printf("Saving pull requests to JSON files...\n")
	fmt.Printf("========================================\n\n")

	// 出力ディレクトリを作成
	outputDir := filepath.Join("data", client.Owner, client.Name)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	totalPRs := len(prs)
	savedCount := 0
	skippedCount := 0
	errorCount := 0
	lastProgressTime := time.Now()

	// 各PRをJSONファイルに保存
	for i, pr := range prs {
		if pr.Number == nil {
			log.Printf("⚠️  Skipping PR with nil number")
			skippedCount++
			continue
		}

		prNumber := *pr.Number
		outputPath := filepath.Join(outputDir, fmt.Sprintf("%d.json", prNumber))

		// PRをJSONにエンコード
		data, err := json.MarshalIndent(pr, "", "  ")
		if err != nil {
			log.Printf("❌ Failed to marshal PR #%d: %v", prNumber, err)
			errorCount++
			continue
		}

		// ファイルに書き込む
		if err := os.WriteFile(outputPath, data, 0644); err != nil {
			log.Printf("❌ Failed to write PR #%d to file: %v", prNumber, err)
			errorCount++
			continue
		}

		savedCount++
		currentProgress := i + 1

		// 10件ごと、または5秒ごとに進捗を表示
		if currentProgress%10 == 0 || time.Since(lastProgressTime) >= 5*time.Second {
			percentage := float64(currentProgress) / float64(totalPRs) * 100
			fmt.Printf("Progress: %d/%d (%.1f%%) - Saved PR #%d\n", currentProgress, totalPRs, percentage, prNumber)
			lastProgressTime = time.Now()
		}
	}

	saveDuration := time.Since(startTime.Add(fetchDuration))
	totalDuration := time.Since(startTime)

	fmt.Printf("\n========================================\n")
	fmt.Printf("Summary\n")
	fmt.Printf("========================================\n")
	fmt.Printf("Total PRs fetched:     %d\n", totalPRs)
	fmt.Printf("Successfully saved:     %d\n", savedCount)
	fmt.Printf("Skipped:                %d\n", skippedCount)
	fmt.Printf("Errors:                 %d\n", errorCount)
	fmt.Printf("Fetch duration:         %v\n", fetchDuration.Round(time.Second))
	fmt.Printf("Save duration:          %v\n", saveDuration.Round(time.Second))
	fmt.Printf("Total duration:         %v\n", totalDuration.Round(time.Second))
	fmt.Printf("Output directory:       %s\n", outputDir)
	fmt.Printf("========================================\n")
	fmt.Printf("✓ Done!\n")
}
