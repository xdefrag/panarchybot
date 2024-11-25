package chatgpt

import (
	"context"

	"github.com/openai/openai-go"
	"github.com/xdefrag/panarchybot/config"
)

type ChatGPT struct {
	cl  *openai.Client
	cfg *config.Config
}

func (c *ChatGPT) MakeQuestion(ctx context.Context, post string) (string, error) {
	comp, err := c.cl.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: openai.F([]openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(c.cfg.OpenAI.Question + post),
		}),
		Model: openai.F(c.cfg.OpenAI.Model),
	})
	if err != nil {
		return "", err
	}

	return comp.Choices[0].Message.Content, nil
}

func New(cl *openai.Client, cfg *config.Config) *ChatGPT {
	return &ChatGPT{cl: cl, cfg: cfg}
}
