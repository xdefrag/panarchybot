package chatgpt

import (
	"context"

	"github.com/davecgh/go-spew/spew"
	"github.com/openai/openai-go"
)

const defaultModel = openai.ChatModelGPT3_5Turbo

type ChatGPT struct {
	cl *openai.Client
}

func (c *ChatGPT) MakeQuestion(ctx context.Context, post string) (string, error) {
	comp, err := c.cl.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: openai.F([]openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(post),
		}),
		Model: openai.F(defaultModel),
	})
	if err != nil {
		return "", err
	}

	spew.Dump(comp)

	return comp.Choices[0].Message.Content, nil
}

func New(cl *openai.Client) *ChatGPT {
	return &ChatGPT{cl: cl}
}
