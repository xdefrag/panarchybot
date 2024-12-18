package tgbot

import (
	"context"
	"log/slog"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/xdefrag/panarchybot"
)

func (t *TGBot) groupHandlerWrapper(next panarchybot.TelegramBotGroupHandler) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, upd *models.Update) {
		l := t.l.With(
			"update_id", upd.ID,
			"channel_id", upd.Message.Chat.ID,
		)

		if err := next(ctx, upd, l); err != nil {
			l.ErrorContext(ctx, "failed to handle message",
				slog.String("error", err.Error()))
			return
		}

		l.DebugContext(ctx, "message handled")
	}
}

func (t *TGBot) messageGroupHandler(ctx context.Context, upd *models.Update, l *slog.Logger) error {
	text := upd.Message.Text

	if text == "" && upd.Message.Caption != "" {
		text = upd.Message.Caption
	}

	res, err := t.gpt.MakeQuestion(ctx, text)
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
					{Text: textFollowUpButton, URL: textFollowUpLink},
				},
			},
		},
		Text: res,
	}); err != nil {
		return err
	}

	return nil
}

const makeAPointCmd = "/make_a_point"

func (t *TGBot) answerGroupHanlder(ctx context.Context, upd *models.Update, l *slog.Logger) error {
	text := upd.Message.ReplyToMessage.Text

	if text == "" && upd.Message.ReplyToMessage.Caption != "" {
		text = upd.Message.ReplyToMessage.Caption
	}

	res, err := t.gpt.MakeQuestion(ctx, text)
	if err != nil {
		return err
	}

	if _, err := t.bot.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: upd.Message.Chat.ID,
		ReplyParameters: &models.ReplyParameters{
			MessageID: upd.Message.ReplyToMessage.ID,
		},
		ReplyMarkup: &models.InlineKeyboardMarkup{
			InlineKeyboard: [][]models.InlineKeyboardButton{
				{
					{Text: textFollowUpButton, URL: textFollowUpLink},
				},
			},
		},
		Text: res,
	}); err != nil {
		return err
	}

	return nil
}
