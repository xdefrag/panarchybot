package tgbot

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/jackc/pgx/v5"
	"github.com/xdefrag/panarchybot/campaign"
	"github.com/xdefrag/panarchybot/chatgpt"
	"github.com/xdefrag/panarchybot/config"
	"github.com/xdefrag/panarchybot/db"
	"github.com/xdefrag/panarchybot/stellar"
)

type TGBot struct {
	cfg      *config.Config
	q        *db.Queries
	bot      *bot.Bot
	stellar  *stellar.Stellar
	gpt      *chatgpt.ChatGPT
	campaign *campaign.Campaign
	l        *slog.Logger
}

func (t *TGBot) Run(ctx context.Context) {
	t.bot.RegisterHandler(bot.HandlerTypeMessageText, "/start", bot.MatchTypeExact,
		t.privateHandlerWrapper(t.startPrivateHandler))
	t.bot.RegisterHandler(bot.HandlerTypeCallbackQueryData, "register", bot.MatchTypeExact,
		t.privateHandlerWrapper(t.registerPrivateHandler))
	t.bot.RegisterHandler(bot.HandlerTypeCallbackQueryData, "start", bot.MatchTypeExact,
		t.privateHandlerWrapper(t.startPrivateHandler))
	t.bot.RegisterHandler(bot.HandlerTypeCallbackQueryData, "suggest", bot.MatchTypePrefix,
		t.privateHandlerWrapper(t.callbackSuggestPrivateHandler))
	t.bot.RegisterHandler(bot.HandlerTypeCallbackQueryData, "send", bot.MatchTypeExact,
		t.privateHandlerWrapper(t.callbackSendPrivateHandler))
	t.bot.RegisterHandler(bot.HandlerTypeCallbackQueryData, "send_confirm", bot.MatchTypeExact,
		t.privateHandlerWrapper(t.callbackSendConfirmPrivateHandler))

	t.bot.RegisterHandlerMatchFunc(func(upd *models.Update) bool {
		return upd.Message != nil && upd.Message.Chat.Type == models.ChatTypePrivate
	}, t.privateHandlerWrapper(t.messagePrivateHandler))

	t.bot.RegisterHandlerMatchFunc(func(upd *models.Update) bool {
		return upd.Message != nil && upd.Message.ForwardOrigin != nil &&
			upd.Message.ForwardOrigin.MessageOriginChannel != nil &&
			upd.Message.ForwardOrigin.MessageOriginChannel.Chat.ID == t.cfg.Telegram.MainChannelID
	}, t.groupHandlerWrapper(t.messageGroupHandler))

	t.bot.RegisterHandlerMatchFunc(func(upd *models.Update) bool {
		return upd.Message != nil && upd.Message.ReplyToMessage != nil &&
			strings.HasPrefix(upd.Message.Text, thanksCmd)
	}, t.groupHandlerWrapper(t.thanksGroupHandler))

	t.bot.Start(ctx)
}

func (t *TGBot) privateHandlerWrapper(next func(ctx context.Context, state db.State, upd *models.Update) error) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, upd *models.Update) {
		if !t.cfg.Telegram.Private.Enable {
			return
		}

		user := getUser(upd)

		l := t.l.With(
			"update_id", upd.ID,
			"user_id", user.ID,
		)

		st, err := t.q.GetState(ctx, user.ID)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			l.ErrorContext(ctx, "failed to get state", slog.String("error", err.Error()))
			return
		}

		if err := next(ctx, st, upd); err != nil {
			l.ErrorContext(ctx, "failed to handle message",
				slog.String("error", err.Error()),
				slog.String("state", st.State))

			_, _ = t.bot.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: user.ID,
				Text:   textError,
			})

			_ = t.startPrivateHandler(ctx, st, upd)
			return
		}

		l.DebugContext(ctx, "message handled",
			slog.String("state", st.State))
	}
}

func getUser(upd *models.Update) models.User {
	if upd.Message != nil && upd.Message.From != nil {
		return *upd.Message.From
	}

	if upd.CallbackQuery != nil {
		return upd.CallbackQuery.From
	}

	return models.User{}
}

func New(
	cfg *config.Config,
	q *db.Queries,
	b *bot.Bot,
	s *stellar.Stellar,
	gpt *chatgpt.ChatGPT,
	campaign *campaign.Campaign,
	l *slog.Logger,
) *TGBot {
	return &TGBot{
		cfg:      cfg,
		bot:      b,
		q:        q,
		stellar:  s,
		gpt:      gpt,
		campaign: campaign,
		l:        l,
	}
}
