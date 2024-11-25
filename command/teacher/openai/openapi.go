package openai

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

var ErrNoResponse = errors.New("no reponse")

type Client struct {
	baseUrl       string
	model         string
	promptContext string
	client        *openai.Client
}

type OptFunc func(store *Client)

func New(logger *slog.Logger, apiKey string, opts ...OptFunc) Client {
	client := Client{
		model:         defaultModel,
		promptContext: defaultContext,
	}

	for _, opt := range opts {
		opt(&client)
	}

	opiOpts := []option.RequestOption{option.WithAPIKey(apiKey)}
	if client.baseUrl != "" {
		opiOpts = append(opiOpts, option.WithBaseURL(client.baseUrl))
	}

	c := openai.NewClient(opiOpts...)
	client.client = c
	return client
}

func (c Client) Prompt(ctx context.Context, prompt string) (string, error) {
	completion, err := c.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: openai.F([]openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(c.promptContext),
			openai.UserMessage(prompt),
		}),
		Model: openai.F(c.model),
	})
	if err != nil {
		return "", fmt.Errorf("error prompting: %v", err)
	}

	if len(completion.Choices) == 0 {
		return "", ErrNoResponse
	}

	return completion.Choices[0].Message.Content, nil
}

func WithBaseUrl(url string) OptFunc {
	return func(client *Client) {
		client.baseUrl = url
	}
}

func WithModel(model string) OptFunc {
	return func(client *Client) {
		client.model = model
	}
}

func WithCustomContext(ctx string) OptFunc {
	return func(client *Client) {
		client.promptContext = ctx
	}
}
