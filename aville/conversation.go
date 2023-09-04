package aville

import (
	"context"
	"github.com/sashabaranov/go-openai"
)

type ConvoGenerator struct {
	client *openai.Client
}

func NewConvo() ConvoGenerator {
	return ConvoGenerator{
		client: openai.NewClient("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"),
	}
}

func (g *ConvoGenerator) GenerateOptionsAndResponses(prompt string) (string, error) {
	resp, err := g.client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
		},
	)
	if err != nil {
		return "", err
	}
	return resp.Choices[0].Message.Content, nil
}
