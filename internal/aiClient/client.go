package aiclient

import (
	"context"
	"strings"
	"time"

	"github.com/JDinABox/sirch/internal/message"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

type Client struct {
	apiClient openai.Client
}

func NewClient(openaiKey string) *Client {
	c := &Client{
		apiClient: openai.NewClient(
			option.WithBaseURL("https://openrouter.ai/api/v1"),
			option.WithAPIKey(openaiKey),
			//option.WithHeader()
			option.WithJSONSet("usage.include", true),
			option.WithHeader("HTTP-Referer", "https://github.com/JDinABox/sirch"),
			option.WithHeader("X-Title", "Sirch"),
		),
	}
	return c
}

func (c *Client) Run(ctx context.Context, q string) (string, error) {
	var sysMsg strings.Builder
	mData := message.MessageData{
		Year:   time.Now().Year(),
		Age:    22,
		Gender: "Male",
	}
	sysMsg.WriteString(message.SystemQueryExpand(3, 5, mData))

	messages := message.AiMessage{
		openai.SystemMessage(sysMsg.String()),
	}

	msgMap, err := message.TemplateToString(message.QueryExpandData, mData)
	if err != nil {
		return "", err
	}
	messages.AddUserAssistantMap(msgMap)

	messages.AddUser(q)
	chatCompletion, err := c.apiClient.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: messages,
		//Model:    "google/gemini-2.5-flash-lite-preview-09-2025",
		Model: "google/gemma-3-12b-it",
	})
	if err != nil {
		return "", err
	}

	return chatCompletion.Choices[0].Message.Content, nil
}
