package aville

import (
	"context"
	"github.com/sashabaranov/go-openai"
	"os"
)

type ConvoGenerator struct {
	client *openai.Client
}

func NewConvo() ConvoGenerator {
	return ConvoGenerator{
		client: openai.NewClient(os.Getenv("OPENAI_API_KEY")),
	}
}

func (g *ConvoGenerator) Generate(prompt string) (string, error) {
	resp, err := g.client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo16K0613,
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
