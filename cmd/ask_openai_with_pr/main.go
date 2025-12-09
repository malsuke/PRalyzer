package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/malsuke/PRalyzer/internal/openai"
)

const (
	processedPRBufferSize = 100
)

// RateLimitError はレート制限エラー（429）を表す
var RateLimitError = errors.New("rate limit exceeded (429)")

func main() {
	if len(os.Args) < 4 {
		log.Fatal("Usage: go run cmd/ask_openai_with_pr/main.go <input-directory> <output-file> <openai-api-key>")
	}

	inputDir := os.Args[1]
	outputFile := os.Args[2]
	openAIAPIKey := os.Args[3]

	client := openai.NewClient(openAIAPIKey)

	indexFile := getIndexFilePath(outputFile)
	if err := initializeFiles(outputFile, indexFile); err != nil {
		log.Fatalf("Failed to initialize files: %v", err)
	}

	processedPRs, err := loadProcessedPRs(indexFile)
	if err != nil {
		log.Fatalf("Failed to load processed PRs: %v", err)
	}

	prBuffer := newProcessedPRBuffer(indexFile)

	if err := processDirectory(inputDir, client, outputFile, processedPRs, prBuffer); err != nil {
		if errors.Is(err, RateLimitError) {
			// 429エラーの場合は処理済みPRを保存してから終了
			if flushErr := prBuffer.flush(); flushErr != nil {
				log.Printf("⚠️  Failed to flush processed PRs: %v", flushErr)
			}
			log.Fatalf("🛑 Rate limit exceeded (429). Processing stopped.\n   Processed PRs have been saved. You can resume later.")
		}
		log.Fatalf("Failed to process directory: %v", err)
	}

	if err := prBuffer.flush(); err != nil {
		log.Fatalf("Failed to flush processed PRs: %v", err)
	}

	count := countProcessedPRs(outputFile)
	fmt.Printf("\n✓ Successfully processed %d PRs\n", count)
	fmt.Printf("✓ Results saved to: %s\n", outputFile)
	fmt.Printf("✓ Index file: %s\n", indexFile)
}

func processDirectory(inputDir string, client *openai.Client, outputFile string, processedPRs map[int]bool, prBuffer *processedPRBuffer) error {
	err := filepath.Walk(inputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !shouldProcessFile(path, info) {
			return nil
		}

		prNumber, err := extractPRNumber(path)
		if err != nil {
			log.Printf("⚠️  Skipping file %s: %v", filepath.Base(path), err)
			return nil
		}

		if processedPRs[prNumber] {
			log.Printf("⏭️  Skipping PR #%d (already processed)", prNumber)
			return nil
		}

		result, err := processPRFile(path, prNumber, client)
		if err != nil {
			if errors.Is(err, RateLimitError) {
				// 429エラーの場合は処理を停止
				log.Printf("🛑 Rate limit exceeded (429) for PR #%d. Stopping processing.", prNumber)
				return err
			}
			log.Printf("⚠️  Failed to process PR #%d: %v", prNumber, err)
			// その他のエラーは空の結果を記録して続行
			result = createEmptyResult(prNumber)
		}

		if err := appendResultJSONL(outputFile, result); err != nil {
			log.Printf("⚠️  Failed to write result for PR #%d: %v", prNumber, err)
			return nil
		}

		processedPRs[prNumber] = true
		if err := prBuffer.add(prNumber); err != nil {
			log.Printf("⚠️  Failed to buffer processed PR #%d: %v", prNumber, err)
		}

		return nil
	})

	return err
}

func shouldProcessFile(path string, info os.FileInfo) bool {
	if info.IsDir() {
		return false
	}

	if !strings.HasSuffix(strings.ToLower(path), ".json") {
		return false
	}

	fileName := filepath.Base(path)
	return !strings.HasPrefix(fileName, ".")
}

func extractPRNumber(filePath string) (int, error) {
	fileName := filepath.Base(filePath)
	prNumberStr := strings.TrimSuffix(fileName, ".json")
	prNumber, err := strconv.Atoi(prNumberStr)
	if err != nil {
		return 0, fmt.Errorf("invalid PR number format: %w", err)
	}
	return prNumber, nil
}

func processPRFile(filePath string, prNumber int, client *openai.Client) (openai.VulnerabilityDetectionResult, error) {
	fmt.Printf("Processing PR #%d: %s\n", prNumber, filePath)

	conversationJSON, err := readAndValidateJSON(filePath)
	if err != nil {
		log.Printf("⚠️  Failed to read/validate JSON for PR #%d: %v", prNumber, err)
		return createEmptyResult(prNumber), nil
	}

	result, err := client.DetectVulnerabilityDiscussion(conversationJSON)
	if err != nil {
		// 429エラーを検出
		if isRateLimitError(err) {
			log.Printf("⚠️  Rate limit exceeded (429) for PR #%d", prNumber)
			return openai.VulnerabilityDetectionResult{}, RateLimitError
		}
		log.Printf("⚠️  Failed to detect vulnerability discussion for PR #%d: %v", prNumber, err)
		return createEmptyResult(prNumber), nil
	}

	fmt.Printf("  ✓ Completed PR #%d\n", prNumber)

	return openai.VulnerabilityDetectionResult{
		PR:                 prNumber,
		RelevantDiscussion: result.RelevantDiscussion,
		Reason:             result.Reason,
	}, nil
}

func readAndValidateJSON(filePath string) ([]byte, error) {
	conversationJSON, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var jsonData interface{}
	if err := json.Unmarshal(conversationJSON, &jsonData); err != nil {
		return nil, fmt.Errorf("invalid JSON format: %w", err)
	}

	return conversationJSON, nil
}

func createEmptyResult(prNumber int) openai.VulnerabilityDetectionResult {
	return openai.VulnerabilityDetectionResult{
		PR:                 prNumber,
		RelevantDiscussion: "",
		Reason:             "",
	}
}

func getIndexFilePath(outputFile string) string {
	dir := filepath.Dir(outputFile)
	base := filepath.Base(outputFile)
	ext := filepath.Ext(base)
	name := strings.TrimSuffix(base, ext)
	return filepath.Join(dir, "."+name+"_index.json")
}

func initializeFiles(outputFile string, indexFile string) error {
	outputDir := filepath.Dir(outputFile)
	if outputDir != "." && outputDir != "" {
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}
	}

	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		file, err := os.Create(outputFile)
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		file.Close()
	}

	return nil
}

func loadProcessedPRs(indexFile string) (map[int]bool, error) {
	processedPRs := make(map[int]bool)

	if _, err := os.Stat(indexFile); os.IsNotExist(err) {
		return processedPRs, nil
	}

	data, err := os.ReadFile(indexFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read index file: %w", err)
	}

	var prs []int
	if err := json.Unmarshal(data, &prs); err != nil {
		return nil, fmt.Errorf("failed to unmarshal index file: %w", err)
	}

	for _, pr := range prs {
		processedPRs[pr] = true
	}

	return processedPRs, nil
}

type processedPRBuffer struct {
	buffer    []int
	indexFile string
}

func newProcessedPRBuffer(indexFile string) *processedPRBuffer {
	return &processedPRBuffer{
		buffer:    make([]int, 0, processedPRBufferSize),
		indexFile: indexFile,
	}
}

func (pb *processedPRBuffer) add(prNumber int) error {
	pb.buffer = append(pb.buffer, prNumber)

	if len(pb.buffer) >= processedPRBufferSize {
		return pb.flush()
	}

	return nil
}

func (pb *processedPRBuffer) flush() error {
	if len(pb.buffer) == 0 {
		return nil
	}

	existingPRs, err := loadProcessedPRs(pb.indexFile)
	if err != nil {
		existingPRs = make(map[int]bool)
	}

	for _, pr := range pb.buffer {
		existingPRs[pr] = true
	}

	prs := make([]int, 0, len(existingPRs))
	for pr := range existingPRs {
		prs = append(prs, pr)
	}

	data, err := json.Marshal(prs)
	if err != nil {
		return fmt.Errorf("failed to marshal processed PRs: %w", err)
	}

	if err := os.WriteFile(pb.indexFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write index file: %w", err)
	}

	pb.buffer = pb.buffer[:0]
	return nil
}

func appendResultJSONL(outputFile string, result openai.VulnerabilityDetectionResult) error {
	file, err := os.OpenFile(outputFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open output file: %w", err)
	}
	defer file.Close()

	jsonData, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("failed to marshal result: %w", err)
	}

	if _, err := file.Write(jsonData); err != nil {
		return fmt.Errorf("failed to write result: %w", err)
	}

	if _, err := file.WriteString("\n"); err != nil {
		return fmt.Errorf("failed to write newline: %w", err)
	}

	return nil
}

func countProcessedPRs(outputFile string) int {
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		return 0
	}

	file, err := os.Open(outputFile)
	if err != nil {
		return 0
	}
	defer file.Close()

	count := 0
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if strings.TrimSpace(scanner.Text()) != "" {
			count++
		}
	}

	return count
}

// isRateLimitError はエラーがレート制限エラー（429）かどうかを判定する
func isRateLimitError(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()

	// HTTPレスポンスエラーの場合
	if strings.Contains(errStr, "429") {
		return true
	}

	// レート制限エラーメッセージをチェック
	if strings.Contains(strings.ToLower(errStr), "rate limit") {
		return true
	}

	return false
}
