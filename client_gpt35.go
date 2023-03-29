package gptest

import (
	"context"
	"fmt"
	"log"
	"os"

	gogpt "github.com/sashabaranov/go-openai"
)

const defaultPrompt = `
Here is a prompt:

You are now a machine helping me write smoke test cases. 
I will send you method declarations one by one, and you will return the corresponding test cases in the relevant language to me. 
No need for explanations. No any extra note.
Always return a code snippet with markdown format only. One response for one case.

Let's start from the next line.
`

type ChatGPTClient struct {
	token  string
	client *gogpt.Client
	topics []string
}

func (c *ChatGPTClient) Prepare(promptFile string) error {
	c.client = gogpt.NewClient(c.token)

	var prompt string
	if promptFile == "" {
		prompt = defaultPrompt
	} else {
		contentBytes, err := os.ReadFile(promptFile)
		if err != nil {
			return err
		}
		prompt = string(contentBytes)
	}

	// prompt
	resp, err := c.client.CreateChatCompletion(context.Background(), gogpt.ChatCompletionRequest{
		Model: gogpt.GPT3Dot5Turbo,
		Messages: []gogpt.ChatCompletionMessage{
			{
				Role:    `system`,
				Content: prompt,
			},
		},
	})
	if err != nil {
		return err
	}
	log.Printf("prompt resp: %v", resp.Choices[0].Message.Content)
	return nil
}

func NewGpt35Client() Client {
	return &ChatGPTClient{}
}

func (c *ChatGPTClient) SetToken(token string) {
	c.token = token
}

func (c *ChatGPTClient) Ask(q string) (string, error) {
	gptClient := gogpt.NewClient(c.token)
	ctx := context.Background()

	resp, err := gptClient.CreateChatCompletion(ctx, gogpt.ChatCompletionRequest{
		Model: gogpt.GPT3Dot5Turbo,
		Messages: []gogpt.ChatCompletionMessage{
			{
				Role: `system`,
				Content: fmt.Sprintf(`
write a unit test case about:

%s
`, q),
			},
		},
	})
	if err != nil {
		return "", err
	}

	return resp.Choices[0].Message.Content, nil
}
