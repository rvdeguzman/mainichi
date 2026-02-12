package adapters

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type chatRequest struct {
	Model    string        `json:"model"`
	Messages []chatMessage `json:"messages"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

func GeneratePrompt(apiKey string) (string, error) {
	req := chatRequest{
		Model: "gpt-4o-mini",
		Messages: []chatMessage{
			{
				Role: "system",
				Content: strings.TrimSpace(`
				You generate ONE journaling prompt for a daily stream-of-consciousness writing ritual.

				Output rules:
				- Output exactly one sentence and nothing else.
				- No preamble, no labels, no quotes, no markdown, no lists.
				- 8–18 words.
				- End with a question mark.

				Style rules:
				- Calm, introspective, concrete, non-preachy.
				- No clichés, no motivational tone, no advice, no therapy language.
				- No “write about…”, “describe…”, “imagine…”, “tell a story…”.
				- Avoid proper nouns, brands, and pop culture references.
				- Avoid heavy topics (self-harm, suicide, abuse, violence).

				Prompt quality:
				- Aim for a gentle steering question that invites honest reflection.
				- Prefer specifics over abstractions (everyday moments, decisions, avoidance, attention, friction, small truths).
			`),
			},
			{
				Role:    "user",
				Content: "Generate today's journaling prompt.",
			},
		},
	}

	body, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}

	var chatResp chatResponse
	if err := json.Unmarshal(respBody, &chatResp); err != nil {
		return "", fmt.Errorf("unmarshal response: %w", err)
	}

	if chatResp.Error != nil {
		return "", fmt.Errorf("openai: %s", chatResp.Error.Message)
	}

	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("openai: no choices returned")
	}

	return strings.TrimSpace(chatResp.Choices[0].Message.Content), nil
}
