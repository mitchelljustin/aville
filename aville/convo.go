package aville

import (
	"context"
	"fmt"
	"github.com/sashabaranov/go-openai"
	"os"
	"strings"
)

type ConvoGenerator struct {
	client *openai.Client
}

func NewConvo() ConvoGenerator {
	return ConvoGenerator{
		client: openai.NewClient(os.Getenv("OPENAI_API_KEY")),
	}
}

func (g *ConvoGenerator) Generate(prompts []string) (string, error) {
	promptMessages := []openai.ChatCompletionMessage{
		{
			Role: openai.ChatMessageRoleSystem,
			Content: `
                        You are generating textual prose content for an interactive visual novel.
                        Number your responses '1.', '2.', '3.', etc.
						When asked to generate player responses, prefix with "Possible responses:".
                    `,
		},
	}
	for _, prompt := range prompts {
		if prompt == "" {
			continue
		}
		promptMessages = append(promptMessages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleUser,
			Content: prompt,
		})
	}
	resp, err := g.client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:    openai.GPT3Dot5Turbo16K,
			Messages: promptMessages,
		},
	)
	if err != nil {
		return "", err
	}
	return resp.Choices[0].Message.Content, nil
}

func extractEntityResponseAndPlayerOptions(input string) (string, string) {
	entityResponse, playerOptions, found := strings.Cut(input, "Possible responses:")
	if found {
		entityResponse = strings.TrimSpace(entityResponse)
		playerOptions = strings.TrimSpace(playerOptions)
		return entityResponse, playerOptions
	}
	lines := strings.Split(input, "\n")
	entityResponseIndex := 0
	for i, line := range lines {
		line := strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if entityResponse == "" {
			entityResponse = line
			entityResponseIndex = i
			break
		}
	}
	if entityResponse == "" { // didn't find it
		return input, ""
	}
	playerOptions = strings.Join(lines[entityResponseIndex+1:], "\n")
	return entityResponse, playerOptions
}

func (m *Model) conductConversation(input string) {
	var convo string
	var err error
	entity := m.interactingEntity
	prompt := entity.Persona
	if input == "" {
		prompt += `
			Say something to me.

			What are three VERY DIFFERENT ways in which I can respond?
		`
	} else {
		m.player.LastResponse = input
		prompt += fmt.Sprintf(`
			You just said to me: %v.

			I said: %v.

			Say something back.
`, entity.LastResponse, input)
	}
	prompt = strings.ReplaceAll(prompt, "\t", "")
	m.displayText(fmt.Sprintf("Generating convo for %v...\nPrompt: \n%v",
		entity.Name, prompt))
	if convo, err = m.convo.Generate([]string{prompt}); err != nil {
		m.displayText(fmt.Sprintf("Error generating convo: \n%v", err))
		return
	}
	fmt.Fprintf(m.logFile, "Prompt: %v\nResponse: %v\n", prompt, convo)
	entity.LastResponse, m.convoOptions = extractEntityResponseAndPlayerOptions(convo)
	var pagerText string
	if input == "" {
		pagerText = fmt.Sprintf("%v. \n\nHow do you respond?\n%v",
			entity.LastResponse, m.convoOptions)
	} else {
		pagerText = entity.LastResponse
	}
	m.displayText(pagerText)
}
