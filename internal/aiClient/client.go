package aiclient

import (
	"context"
	"encoding/json/v2"
	"strings"
	"time"

	"github.com/JDinABox/sirch/internal/message"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

type Client struct {
	apiClient openai.Client
}
type CompletionUsage struct {
	Cost        float64 `json:"cost"`
	IsByoK      bool    `json:"is_byok"`
	CostDetails struct {
		UpstreamInferenceCost            float64 `json:"upstream_inference_cost"`
		UpstreamInferencePromptCost      float64 `json:"upstream_inference_prompt_cost"`
		UpstreamInferenceCompletionsCost float64 `json:"upstream_inference_completions_cost"`
	} `json:"cost_details"`
}

type Output struct {
	Content string  `json:"content"`
	Cost    float64 `json:"cost"`
}

func NewClient(baseurl, openaiKey string) *Client {
	c := &Client{
		apiClient: openai.NewClient(
			option.WithBaseURL(baseurl),
			option.WithAPIKey(openaiKey),
			//option.WithHeader()
			option.WithJSONSet("usage.include", true),
			option.WithHeader("HTTP-Referer", "https://github.com/JDinABox/sirch"),
			option.WithHeader("X-Title", "Sirch"),
		),
	}
	return c
}

func (c *Client) RunQueryExpand(ctx context.Context, q string) (Output, error) {
	var sysMsg strings.Builder
	mData := message.MessageData{
		Year:   time.Now().Year(),
		Age:    22,
		Gender: "Male",
	}
	sysMsg.WriteString(message.SystemQueryExpand(3, 5, mData))

	us, err := message.TemplateToUserAssistant(message.QueryExpandData, mData)
	if err != nil {
		return Output{}, err
	}

	return c.Run(ctx, "google/gemma-3-12b-it", sysMsg.String(), q, us...)
}

func (c *Client) Run(ctx context.Context, model, system, query string, messages ...message.UserAssistant) (Output, error) {
	m := message.AiMessage{
		openai.SystemMessage(system),
	}
	if len(messages) > 0 {
		m.AddUserAssistant(messages)
	}
	m.AddUser(query)
	chatCompletion, err := c.apiClient.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: m,
		//Model:    "google/gemini-2.5-flash-lite-preview-09-2025",
		Model: model,
		//ReasoningEffort: "minimal",
	})
	if err != nil {
		return Output{}, err
	}

	cost := CompletionUsage{}
	if err = json.Unmarshal([]byte(chatCompletion.Usage.RawJSON()), &cost); err != nil {
		return Output{}, err
	}

	return Output{Content: chatCompletion.Choices[0].Message.Content, Cost: cost.Cost}, nil
}
