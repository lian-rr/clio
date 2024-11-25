package openai

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

// ErrNoResponse thrown when there is no response from the open ai API.
var ErrNoResponse = errors.New("no reponse")

// Client holds the OpenAI client and some configuration.
type Client struct {
	baseUrl       string
	model         string
	promptContext string
	client        *openai.Client
}

// OptFunc used for setting optional configs.
type OptFunc func(store *Client)

// New  returns a new OpenAI client.
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

// Prompt executes a prompt to the OpenAI endpoints
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

// WithBaseUrl sets the Client baseUrl optional param.
func WithBaseUrl(url string) OptFunc {
	return func(client *Client) {
		client.baseUrl = url
	}
}

// WithBaseUrl sets the Client model optional param.
func WithModel(model string) OptFunc {
	return func(client *Client) {
		client.model = model
	}
}

// WithBaseUrl sets the Client prompt context optional param.
func WithCustomContext(ctx string) OptFunc {
	return func(client *Client) {
		client.promptContext = ctx
	}
}
