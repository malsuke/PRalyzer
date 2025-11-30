package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	gh "github.com/google/go-github/v77/github"
	"github.com/malsuke/PRalyzer/internal/github"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: PRalyzer <repository-url> [github-pat]\nNote: GitHub PAT is optional but recommended to avoid rate limiting")
	}

	repoURL := os.Args[1]
	var githubPAT string
	if len(os.Args) >= 3 {
		githubPAT = os.Args[2]
	} else {
		fmt.Println("Warning: No GitHub PAT provided. Rate limiting may occur.")
	}

	client, err := github.NewClient(githubPAT, repoURL, nil)
	if err != nil {
		log.Fatalf("Failed to create GitHub client: %v", err)
	}

	words, err := loadWordList("word_list.json")
	if err != nil {
		log.Fatalf("Failed to load word list: %v", err)
	}

	// ベースデータディレクトリ
	baseDataDir := filepath.Join("data", client.Owner, client.Name)

	// 処理済みPR番号を読み込む
	processedPRsFile := filepath.Join(baseDataDir, ".processed_prs.json")
	processedPRs, err := loadProcessedPRs(processedPRsFile)
	if err != nil {
		log.Printf("Failed to load processed PRs (will start fresh): %v", err)
		processedPRs = make(map[int]bool)
	} else {
		fmt.Printf("Loaded %d previously processed PRs\n", len(processedPRs))
	}

	// キーワードごとに処理
	for _, word := range words {
		fmt.Printf("\n=== Processing keyword: %s ===\n", word)

		// 1. キーワードでPRを検索
		prs, err := client.SearchPullRequestsWithCommentKeyword(word)
		if err != nil {
			if isRateLimitError(err) {
				log.Printf("Rate limit exceeded while searching PRs with keyword '%s'. Stopping.", word)
				log.Printf("Tip: Provide a GitHub PAT to increase rate limits.")
				break
			}
			log.Printf("Failed to search PRs with keyword '%s': %v", word, err)
			continue
		}

		if len(prs) == 0 {
			fmt.Printf("No PRs found for keyword '%s'. Skipping.\n", word)
			continue
		}

		fmt.Printf("Found %d PRs for keyword '%s'\n", len(prs), word)

		// 2. キーワードごとのディレクトリを作成
		keywordDir := filepath.Join(baseDataDir, word)
		if err := os.MkdirAll(keywordDir, 0755); err != nil {
			log.Printf("Failed to create directory for keyword '%s': %v", word, err)
			continue
		}

		// 3. 各PRのコメントを取得してファイルに保存
		processedInThisKeyword := 0
		for _, pr := range prs {
			if pr.Number == nil {
				continue
			}

			prNumber := *pr.Number

			// 既に処理済みのPRはスキップ
			if processedPRs[prNumber] {
				fmt.Printf("Skipping PR #%d (already processed)\n", prNumber)
				continue
			}

			fmt.Printf("Fetching comments for PR #%d\n", prNumber)

			comments, err := client.GetComments(prNumber)
			if err != nil {
				if isRateLimitError(err) {
					log.Printf("Rate limit exceeded while fetching comments for PR #%d. Stopping keyword '%s'.", prNumber, word)
					log.Printf("Tip: Provide a GitHub PAT to increase rate limits.")
					// 処理済みPR番号を保存してから終了
					if err := saveProcessedPRs(processedPRsFile, processedPRs); err != nil {
						log.Printf("Failed to save processed PRs: %v", err)
					}
					return
				}
				log.Printf("Failed to get comments for PR #%d: %v", prNumber, err)
				continue
			}

			// JSONファイルに書き込む
			outputPath := filepath.Join(keywordDir, fmt.Sprintf("%d.json", prNumber))
			if err := writeCommentsToFile(comments, outputPath); err != nil {
				log.Printf("Failed to write comments to file for PR #%d: %v", prNumber, err)
				continue
			}

			// 処理済みとしてマーク
			processedPRs[prNumber] = true
			processedInThisKeyword++

			fmt.Printf("Saved comments for PR #%d to %s\n", prNumber, outputPath)

			// 10件処理するごとに処理済みPR番号を保存（進捗を保存）
			if processedInThisKeyword%10 == 0 {
				if err := saveProcessedPRs(processedPRsFile, processedPRs); err != nil {
					log.Printf("Failed to save processed PRs: %v", err)
				}
			}
		}

		// キーワードごとの処理が完了したら処理済みPR番号を保存
		if processedInThisKeyword > 0 {
			if err := saveProcessedPRs(processedPRsFile, processedPRs); err != nil {
				log.Printf("Failed to save processed PRs: %v", err)
			}
		}
	}

	// 最終的に処理済みPR番号を保存
	if err := saveProcessedPRs(processedPRsFile, processedPRs); err != nil {
		log.Printf("Failed to save processed PRs: %v", err)
	}

	fmt.Println("\nDone!")
}

// loadWordList はword_list.jsonファイルを読み込む
func loadWordList(filename string) ([]string, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read word list file: %w", err)
	}

	var words []string
	if err := json.Unmarshal(data, &words); err != nil {
		return nil, fmt.Errorf("failed to parse word list JSON: %w", err)
	}

	return words, nil
}

// writeCommentsToFile はコメントをJSONファイルに書き込む
func writeCommentsToFile(comments []*gh.IssueComment, filepath string) error {
	data, err := json.MarshalIndent(comments, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal comments: %w", err)
	}

	if err := os.WriteFile(filepath, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// isRateLimitError はエラーがレート制限エラーかどうかを判定する
func isRateLimitError(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()

	// HTTPレスポンスエラーの場合
	if strings.Contains(errStr, "403") || strings.Contains(errStr, "429") {
		return true
	}

	// go-githubのレート制限エラーメッセージをチェック
	if strings.Contains(strings.ToLower(errStr), "rate limit") {
		return true
	}

	// RateLimitError型をチェック
	if rateLimitErr, ok := err.(*gh.RateLimitError); ok {
		return rateLimitErr != nil
	}

	// AbuseRateLimitError型をチェック
	if abuseRateLimitErr, ok := err.(*gh.AbuseRateLimitError); ok {
		return abuseRateLimitErr != nil
	}

	return false
}

// waitForRateLimitReset はレート制限がリセットされるまで待機する（オプション機能）
func waitForRateLimitReset(err error) {
	if rateLimitErr, ok := err.(*gh.RateLimitError); ok {
		resetTime := rateLimitErr.Rate.Reset.Time
		waitDuration := time.Until(resetTime)
		if waitDuration > 0 {
			log.Printf("Waiting %v for rate limit reset...", waitDuration)
			time.Sleep(waitDuration)
		}
	}
}

// loadProcessedPRs は処理済みPR番号をファイルから読み込む
func loadProcessedPRs(filepath string) (map[int]bool, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		if os.IsNotExist(err) {
			// ファイルが存在しない場合は空のマップを返す
			return make(map[int]bool), nil
		}
		return nil, fmt.Errorf("failed to read processed PRs file: %w", err)
	}

	var prNumbers []int
	if err := json.Unmarshal(data, &prNumbers); err != nil {
		return nil, fmt.Errorf("failed to parse processed PRs JSON: %w", err)
	}

	processedPRs := make(map[int]bool)
	for _, prNumber := range prNumbers {
		processedPRs[prNumber] = true
	}

	return processedPRs, nil
}

// saveProcessedPRs は処理済みPR番号をファイルに保存する
func saveProcessedPRs(filePath string, processedPRs map[int]bool) error {
	// マップをスライスに変換
	prNumbers := make([]int, 0, len(processedPRs))
	for prNumber := range processedPRs {
		prNumbers = append(prNumbers, prNumber)
	}

	data, err := json.MarshalIndent(prNumbers, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal processed PRs: %w", err)
	}

	// ディレクトリが存在することを確認
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}
