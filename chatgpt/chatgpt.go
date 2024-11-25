package chatgpt

import (
	"context"

	"github.com/openai/openai-go"
)

const defaultModel = openai.ChatModelGPT4o

type ChatGPT struct {
	cl *openai.Client
}

func (c *ChatGPT) MakeQuestion(ctx context.Context, post string) (string, error) {
	comp, err := c.cl.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: openai.F([]openai.ChatCompletionMessageParamUnion{
			openai.UserMessage("Ты саркастичный комментатор, помощник автора поста, предерживаешься крайних анархо-капиталистических взглядов. Тебе необходимо написать очень краткое мнение о посте и задать вопрос публике. Разбивай написанное на абзацы, чтобы было удобно читать. Объем не должен привышать 80 символов. Обязательно призвать нажимать на кнопку под этим постом, чтобы получить эйрдроп. Необходимо максимально интегрировать просьбу в мнение. Сам пост:\n" + post),
		}),
		Model: openai.F(defaultModel),
	})
	if err != nil {
		return "", err
	}

	return comp.Choices[0].Message.Content, nil
}

func New(cl *openai.Client) *ChatGPT {
	return &ChatGPT{cl: cl}
}
