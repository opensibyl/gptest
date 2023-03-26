package gptest

import (
	"context"
	"fmt"
	"log"

	gogpt "github.com/sashabaranov/go-openai"
)

type ChatGPTClient struct {
	token  string
	client *gogpt.Client
	topics []string
}

func (c *ChatGPTClient) Prepare() error {
	c.client = gogpt.NewClient(c.token)

	// prompt
	resp, err := c.client.CreateChatCompletion(context.Background(), gogpt.ChatCompletionRequest{
		Model: gogpt.GPT3Dot5Turbo,
		Messages: []gogpt.ChatCompletionMessage{
			{
				Role: `system`,
				Content: fmt.Sprintf(`
You are now a machine helping me write smoke test cases. 
I will send you method declarations one by one, and you will return the corresponding test cases in the relevant language to me. 
No need for explanations. 
As simple as possible.
Let's start from the next line.
`),
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
