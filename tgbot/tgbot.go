package tgbot

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/jackc/pgx/v5"
	"github.com/xdefrag/panarchybot"
	"github.com/xdefrag/panarchybot/campaign"
	"github.com/xdefrag/panarchybot/chatgpt"
	"github.com/xdefrag/panarchybot/config"
	"github.com/xdefrag/panarchybot/db"
	"github.com/xdefrag/panarchybot/metrics"
)

type TGBot struct {
	cfg      *config.Config
	q        *db.Queries
	bot      *bot.Bot
	ledger   panarchybot.Ledger
	gpt      *chatgpt.ChatGPT
	campaign *campaign.Campaign
	l        *slog.Logger
}

func (t *TGBot) Run(ctx context.Context) {
	privateMWs := []panarchybot.TelegramBotPrivateMiddleware{
		metrics.TelegramBotPrivateMiddleware,
	}

	t.bot.RegisterHandler(bot.HandlerTypeMessageText, "/start", bot.MatchTypeExact,
		t.newPrivateHandler(t.startPrivateHandler, privateMWs...))
	t.bot.RegisterHandler(bot.HandlerTypeCallbackQueryData, "register", bot.MatchTypeExact,
		t.newPrivateHandler(t.registerPrivateHandler, privateMWs...))
	t.bot.RegisterHandler(bot.HandlerTypeCallbackQueryData, "start", bot.MatchTypeExact,
		t.newPrivateHandler(t.startPrivateHandler, privateMWs...))
	t.bot.RegisterHandler(bot.HandlerTypeCallbackQueryData, "suggest", bot.MatchTypePrefix,
		t.newPrivateHandler(t.callbackSuggestPrivateHandler, privateMWs...))
	t.bot.RegisterHandler(bot.HandlerTypeCallbackQueryData, "send", bot.MatchTypeExact,
		t.newPrivateHandler(t.callbackSendPrivateHandler, privateMWs...))
	t.bot.RegisterHandler(bot.HandlerTypeCallbackQueryData, "send_confirm", bot.MatchTypeExact,
		t.newPrivateHandler(t.callbackSendConfirmPrivateHandler, privateMWs...))

	t.bot.RegisterHandlerMatchFunc(func(upd *models.Update) bool {
		return upd.Message != nil && upd.Message.Chat.Type == models.ChatTypePrivate
	}, t.newPrivateHandler(t.messagePrivateHandler, privateMWs...))

	groupMWs := []panarchybot.TelegramBotGroupMiddleware{
		metrics.TelegramBotGroupMiddleware,
	}

	t.bot.RegisterHandlerMatchFunc(func(upd *models.Update) bool {
		return upd.Message != nil && upd.Message.ForwardOrigin != nil &&
			upd.Message.ForwardOrigin.MessageOriginChannel != nil &&
			upd.Message.ForwardOrigin.MessageOriginChannel.Chat.ID == t.cfg.Telegram.MainChannelID
	}, t.newGroupHandler(t.messageGroupHandler, groupMWs...))

	t.bot.RegisterHandlerMatchFunc(func(upd *models.Update) bool {
		return upd.Message != nil && upd.Message.ReplyToMessage != nil &&
			strings.HasPrefix(upd.Message.Text, thanksCmd)
	}, t.newGroupHandler(t.thanksGroupHandler, groupMWs...))

	t.bot.Start(ctx)
}

func (t *TGBot) newPrivateHandler(handler panarchybot.TelegramBotPrivateHandler, mws ...panarchybot.TelegramBotPrivateMiddleware) bot.HandlerFunc {
	for _, mw := range mws {
		handler = mw(handler)
	}
	return t.privateHandlerWrapper(handler)
}

func (t *TGBot) newGroupHandler(handler panarchybot.TelegramBotGroupHandler, mws ...panarchybot.TelegramBotGroupMiddleware) bot.HandlerFunc {
	for _, mw := range mws {
		handler = mw(handler)
	}
	return t.groupHandlerWrapper(handler)
}

func (t *TGBot) privateHandlerWrapper(next panarchybot.TelegramBotPrivateHandler) bot.HandlerFunc {
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

		if err := next(ctx, st, upd, l); err != nil {
			l.ErrorContext(ctx, "failed to handle message",
				slog.String("error", err.Error()),
				slog.String("state", st.State))

			_, _ = t.bot.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: user.ID,
				Text:   textError,
			})

			_ = t.startPrivateHandler(ctx, st, upd, l)
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
	ledger panarchybot.Ledger,
	gpt *chatgpt.ChatGPT,
	campaign *campaign.Campaign,
	l *slog.Logger,
) *TGBot {
	return &TGBot{
		cfg:      cfg,
		bot:      b,
		q:        q,
		ledger:   ledger,
		gpt:      gpt,
		campaign: campaign,
		l:        l,
	}
}
