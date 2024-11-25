package tgbot

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jackc/pgx/v5"
	"github.com/looplab/fsm"
	"github.com/xdefrag/panarchybot/chatgpt"
	"github.com/xdefrag/panarchybot/config"
	"github.com/xdefrag/panarchybot/db"
)

type TGBot struct {
	l   *slog.Logger
	cfg *config.Config
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

			if err := t.handle(ctx, upd); err != nil {
				l.ErrorContext(ctx, "failed to handle update",
					slog.String("error", err.Error()))
			}
		case <-ctx.Done():
			return
		}
	}
}

func (t *TGBot) handle(ctx context.Context, upd tgbotapi.Update) error {
	if upd.Message != nil && upd.Message.ForwardFromChat != nil &&
		upd.Message.ForwardFromChat.ID == t.cfg.Telegram.MainChannelID {
		return t.handleMainChannel(ctx, upd)
	}

	if upd.Message != nil && upd.Message.Chat != nil && upd.Message.Chat.IsPrivate() {
		return t.handlePrivate(ctx, upd)
	}

	return nil
}

func (t *TGBot) handleMainChannel(ctx context.Context, upd tgbotapi.Update) error {
	q, err := t.gpt.MakeQuestion(ctx, upd.Message.Text)
	if err != nil {
		return err
	}

	msg := tgbotapi.NewMessage(upd.Message.Chat.ID, q)
	msg.ReplyToMessageID = upd.Message.MessageID
	msg.ParseMode = tgbotapi.ModeHTML
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL(t.cfg.Telegram.FollowUp.Message, t.cfg.Telegram.FollowUp.URL),
		),
	)
	if _, err := t.bot.Send(msg); err != nil {
		return err
	}

	return nil
}

func (t *TGBot) handlePrivate(ctx context.Context, upd tgbotapi.Update) error {
	st, err := t.q.GetState(ctx, upd.Message.From.ID)
	if errors.Is(err, pgx.ErrNoRows) {
		st = db.State{
			UserID: upd.Message.From.ID,
			State:  stateInit,
			Data:   make(map[string]interface{}),
			Meta:   make(map[string]interface{}),
		}
	}
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil
	}

	sm := t.getSM()
	sm.SetState(st.State)

	st.Data["message"] = upd.Message.Text
	st.Data["message_id"] = upd.Message.MessageID
	st.Meta["username"] = upd.Message.From.UserName
	st.Meta["firstname"] = upd.Message.From.FirstName
	st.Meta["lastname"] = upd.Message.From.LastName
	st.Meta["chat_type"] = upd.Message.Chat.Type
	st.Meta["chat_title"] = upd.Message.Chat.Title
	st.Meta["chat_id"] = upd.Message.Chat.ID

	ev, args := prepareEventAndArgs(upd.Message.Text, st)

	if sm.Is(stateSuggest) {
		ev = eventSuggested
	}

	if trueEvent, ok := mapButtonEvent[ev]; ok {
		ev = trueEvent
	}

	if err := sm.Event(ctx, ev, args...); err != nil && !errors.Is(err, fsm.NoTransitionError{}) {
		return err
	}

	return nil
}

var eventsWithArgs = []string{}

func prepareEventAndArgs(text string, args ...interface{}) (string, []interface{}) {
	ev := text

	for _, e := range eventsWithArgs {
		if strings.HasPrefix(text, e) {
			id := text[len(e)+1:]
			args = append(args, id)
			ev = e
		}
	}

	return ev, args
}

func New(
	l *slog.Logger,
	cfg *config.Config,
	q *db.Queries,
	bot *tgbotapi.BotAPI,
	gpt *chatgpt.ChatGPT,
) *TGBot {
	return &TGBot{
		l:   l,
		cfg: cfg,
		q:   q,
		bot: bot,
		gpt: gpt,
	}
}
