package openai

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/shared"
)

type Client struct {
	client openai.Client
}

func NewClient(apiKey string) *Client {
	client := openai.NewClient(
		option.WithAPIKey(apiKey),
	)
	return &Client{
		client: client,
	}
}

func (c *Client) DetectVulnerabilityDiscussion(conversationJSON []byte) (*VulnerabilityDetectionResponse, error) {
	prompt := c.buildPrompt(conversationJSON)

	responseFormat := openai.ChatCompletionNewParamsResponseFormatUnion{
		OfJSONObject: &shared.ResponseFormatJSONObjectParam{},
	}

	chatCompletion, err := c.client.Chat.Completions.New(context.Background(), openai.ChatCompletionNewParams{
		Model: openai.ChatModelGPT5Mini,
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage("Analyze code review discussions for security vulnerability findings. Return JSON only."),
			openai.UserMessage(prompt),
		},
		ResponseFormat: responseFormat,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create chat completion: %w", err)
	}

	if len(chatCompletion.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	content := chatCompletion.Choices[0].Message.Content
	if content == "" {
		return nil, fmt.Errorf("empty content in response")
	}

	var result VulnerabilityDetectionResponse
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON response: %w", err)
	}

	return &result, nil
}

func (c *Client) buildPrompt(conversationJSON []byte) string {
	return fmt.Sprintf(`Analyze this code review conversation for security vulnerability findings.

Conversation:
%s

Return JSON:
{
  "relevant_discussion": "excerpt if vulnerability found, else empty string",
  "reason": "explanation in Japanese if found, else empty string"
}`, string(conversationJSON))
}
