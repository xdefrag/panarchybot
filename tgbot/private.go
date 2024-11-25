package tgbot

import (
	"context"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/looplab/fsm"
	"github.com/xdefrag/panarchybot/db"
)

const (
	stateInit    = "state_init"
	stateStart   = "state_start"
	stateSuggest = "state_suggest"

	eventStart     = "/start"
	eventSuggest   = "/suggest"
	eventSuggested = "/suggested"

	buttonStart   = "💰 Баланс"
	buttonSuggest = "🧌 Предложить пост"
)

var mapButtonEvent = map[string]string{
	buttonStart:   eventStart,
	buttonSuggest: eventSuggest,
}

var buttonsStart = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton(buttonStart),
		tgbotapi.NewKeyboardButton(buttonSuggest),
	),
)

func (t *TGBot) getSM() *fsm.FSM {
	return fsm.NewFSM(
		stateInit,
		fsm.Events{
			{Name: eventStart, Src: []string{stateInit, stateStart}, Dst: stateStart},
			{Name: eventSuggest, Src: []string{stateStart}, Dst: stateSuggest},
			{Name: eventSuggested, Src: []string{stateSuggest}, Dst: stateStart},
		},
		fsm.Callbacks{
			"before_event": func(ctx context.Context, e *fsm.Event) {
				st := e.Args[0].(db.State)
				if err := t.q.CreateState(ctx, db.CreateStateParams{
					UserID: st.UserID,
					State:  e.Dst,
					Data:   st.Data,
					Meta:   st.Meta,
				}); err != nil {
					e.Cancel(err)
					return
				}
			},
			eventStart: func(ctx context.Context, e *fsm.Event) {
				st := e.Args[0].(db.State)

				msg := tgbotapi.NewMessage(st.UserID, "")
				msg.ParseMode = tgbotapi.ModeHTML
				msg.ReplyMarkup = buttonsStart
				if _, err := t.bot.Send(msg); err != nil {
					e.Cancel(err)
					return
				}
			},
			eventSuggest: func(ctx context.Context, e *fsm.Event) {
				st := e.Args[0].(db.State)
				msg := tgbotapi.NewMessage(st.UserID, t.cfg.Telegram.Suggest.SuggestMessage)
				msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
				if _, err := t.bot.Send(msg); err != nil {
					e.Cancel(err)
					return
				}
			},
			eventSuggested: func(ctx context.Context, e *fsm.Event) {
				st := e.Args[0].(db.State)

				fw := tgbotapi.NewForward(t.cfg.Telegram.SuggestChatID, st.Meta["chat_id"].(int64), st.Data["message_id"].(int))
				if _, err := t.bot.Send(fw); err != nil {
					e.Cancel(err)
					return
				}

				msg := tgbotapi.NewMessage(st.UserID, t.cfg.Telegram.Suggest.SuggestedMessage)
				msg.ReplyMarkup = buttonsStart
				if _, err := t.bot.Send(msg); err != nil {
					e.Cancel(err)
					return
				}
			},
		},
	)
}
