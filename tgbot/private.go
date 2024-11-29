package tgbot

import (
	"context"
	"errors"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jackc/pgx/v5"
	"github.com/looplab/fsm"
	"github.com/xdefrag/panarchybot/db"
)

const (
	stateInit        = "state_init"
	stateStart       = "state_start"
	stateSend        = "state_send"
	stateSendTo      = "state_send_to"
	stateSendAmount  = "state_send_amount"
	stateSendConfirm = "state_send_confirm"
	stateSuggest     = "state_suggest"

	eventStart     = "/start"
	eventRegister  = "/register"
	eventSuggest   = "/suggest"
	eventSuggested = "/suggested"
	eventSend      = "/send"
	eventGet       = "/get"

	buttonRegister = "📝 Регистрация"
	buttonStart    = "💰 Баланс"
	buttonSend     = "📤 Отправить"
	buttonSuggest  = "🧌 Предложить пост"
)

var mapButtonEvent = map[string]string{
	buttonRegister: eventRegister,
	buttonStart:    eventStart,
	buttonSend:     eventSend,
	buttonSuggest:  eventSuggest,
}

var buttonsStart = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton(buttonStart),
		tgbotapi.NewKeyboardButton(buttonSend),
	),
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton(buttonSuggest),
	),
)

var buttonsStartNewbie = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton(buttonRegister),
	),
)

func (t *TGBot) getSM() *fsm.FSM {
	return fsm.NewFSM(
		stateInit,
		fsm.Events{
			{Name: eventStart, Src: []string{stateInit, stateStart}, Dst: stateStart},
			{Name: eventSuggest, Src: []string{stateStart}, Dst: stateSuggest},
			{Name: eventSuggested, Src: []string{stateSuggest}, Dst: stateStart},
			{Name: eventRegister, Src: []string{stateStart}, Dst: stateStart},
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

				acc, err := t.q.GetAccount(ctx, st.UserID)
				if err != nil && !errors.Is(err, pgx.ErrNoRows) {
					e.Cancel(err)
					return
				}

				if errors.Is(err, pgx.ErrNoRows) {
					msg := tgbotapi.NewMessage(st.UserID, textNewbie)
					msg.ReplyMarkup = buttonsStartNewbie
					if _, err := t.bot.Send(msg); err != nil {
						e.Cancel(err)
						return
					}
					return
				}

				bal, err := t.stellar.GetBalance(ctx, acc.Address)
				if err != nil {
					e.Cancel(err)
					return
				}

				msg := tgbotapi.NewMessage(st.UserID, fmt.Sprintf(textStart, bal))
				msg.ParseMode = tgbotapi.ModeHTML
				msg.ReplyMarkup = buttonsStart
				if _, err := t.bot.Send(msg); err != nil {
					e.Cancel(err)
					return
				}
			},
			eventRegister: func(ctx context.Context, e *fsm.Event) {
				st := e.Args[0].(db.State)

				pair, err := t.stellar.CreateAccount(ctx)
				if err != nil {
					e.Cancel(err)
					return
				}

				num, err := t.q.CreateAccount(ctx, db.CreateAccountParams{
					UserID:  st.UserID,
					Address: pair.Address(),
					Seed:    pair.Seed(),
				})
				if err != nil {
					e.Cancel(err)
					return
				}

				_ = num // TODO airdrop

				msg := tgbotapi.NewMessage(st.UserID, textRegistered)
				msg.ReplyMarkup = buttonsStart
				msg.ParseMode = tgbotapi.ModeHTML
				if _, err := t.bot.Send(msg); err != nil {
					e.Cancel(err)
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
