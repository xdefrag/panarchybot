package tgbot

import (
	"context"
	"log/slog"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/xdefrag/panarchybot/chatgpt"
	"github.com/xdefrag/panarchybot/db"
)

type TGBot struct {
	l   *slog.Logger
	q   *db.Queries
	bot *tgbotapi.BotAPI
	gpt *chatgpt.ChatGPT
}

func (t *TGBot) Run(ctx context.Context) {
	updchan := t.bot.GetUpdatesChan(tgbotapi.NewUpdate(0))

	for {
		select {
		case upd := <-updchan:
			l := t.l.
				WithGroup("tgbot").
				With(slog.Int("update_id", upd.UpdateID))

			l.DebugContext(ctx, "new message")

			if err := t.handle(ctx, upd, l); err != nil {
				l.ErrorContext(ctx, "failed to handle update",
					slog.String("error", err.Error()))
			}
		case <-ctx.Done():
			return
		}
	}
}

const (
	mainChannelID = -1001892370893
)

func (t *TGBot) handle(ctx context.Context, upd tgbotapi.Update, l *slog.Logger) error {
	if upd.Message != nil && upd.Message.ForwardFromChat != nil && upd.Message.ForwardFromChat.ID == mainChannelID {
		return t.handleMainChannel(ctx, upd, l)
	}

	return nil
}

func (t *TGBot) handleMainChannel(ctx context.Context, upd tgbotapi.Update, l *slog.Logger) error {
	l.InfoContext(ctx, "main channel message")

	q, err := t.gpt.MakeQuestion(ctx, upd.Message.Text)
	if err != nil {
		return err
	}

	msg := tgbotapi.NewMessage(upd.Message.Chat.ID, q)
	msg.ReplyToMessageID = upd.Message.MessageID
	msg.ParseMode = tgbotapi.ModeHTML
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL("Запишись на эйрдроп", "https://panarchybot.t.me"),
		),
	)
	if _, err := t.bot.Send(msg); err != nil {
		return err
	}

	return nil
}

func New(
	l *slog.Logger,
	q *db.Queries,
	bot *tgbotapi.BotAPI,
	gpt *chatgpt.ChatGPT,
) *TGBot {
	return &TGBot{
		l:   l,
		q:   q,
		bot: bot,
		gpt: gpt,
	}
}
