package tgbot

import (
	"context"
	"log/slog"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func (t *TGBot) groupHandlerWrapper(next func(ctx context.Context, upd *models.Update) error) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, upd *models.Update) {
		l := t.l.With(
			"update_id", upd.ID,
			"channel_id", upd.Message.Chat.ID,
		)

		if err := next(ctx, upd); err != nil {
			l.ErrorContext(ctx, "failed to handle message",
				slog.String("error", err.Error()))
			return
		}

		l.DebugContext(ctx, "message handled")
	}
}

func (t *TGBot) messageGroupHandler(ctx context.Context, upd *models.Update) error {
	res, err := t.gpt.MakeQuestion(ctx, upd.Message.Text)
	if err != nil {
		return err
	}

	if _, err := t.bot.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: upd.Message.Chat.ID,
		ReplyParameters: &models.ReplyParameters{
			MessageID: upd.Message.ID,
		},
		ReplyMarkup: &models.InlineKeyboardMarkup{
			InlineKeyboard: [][]models.InlineKeyboardButton{
				{
					{Text: "Забирай эйрдроп", URL: "https://panarchybot.t.me"},
				},
			},
		},
		Text: res,
	}); err != nil {
		return err
	}

	return nil
}
