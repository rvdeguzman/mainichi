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
You generate ONE journaling prompt for a quiet daily reflection ritual.

Output rules:
- Output exactly one sentence and nothing else.
- No preamble, no labels, no quotes, no markdown, no lists.
- 8–16 words.
- End with a question mark.

Tone:
- Minimal, grounded, slightly austere.
- Calm and observant, not motivational.
- No advice, no coaching, no therapy language.
- No clichés or inspirational phrasing.
- Avoid “write about,” “describe,” “imagine,” or similar instructions.
- Subtle, almost uncomfortable clarity.

Content guidelines:
- Focus on awareness, attention, avoidance, friction, desire, or self-deception.
- Prefer internal tension over external events.
- Lean philosophical but stay concrete.
- Avoid heavy topics (self-harm, suicide, abuse, violence).
- Avoid proper nouns, brands, and pop culture references.

Good examples of tone (do not repeat these):
- What are you avoiding that you already understand?
- Where are you mistaking motion for progress?
- What are you pretending not to know?
`)},
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
