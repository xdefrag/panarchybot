package tgbot

import (
	"context"
	"errors"
	"fmt"

	"github.com/davecgh/go-spew/spew"
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

	eventStart       = "/start"
	eventRegister    = "/register"
	eventSuggest     = "/suggest"
	eventSuggested   = "/suggested"
	eventSend        = "/send"
	eventSendTo      = "/send_to"
	eventSendAmount  = "/send_amount"
	eventSendConfirm = "/send_confirm"
	eventGet         = "/get"

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

var waitForInputState = map[string]string{
	stateSuggest:     eventSuggested,
	stateSendTo:      eventSendTo,
	stateSendAmount:  eventSendAmount,
	stateSendConfirm: eventSendConfirm,
}

func (t *TGBot) getSM() *fsm.FSM {
	return fsm.NewFSM(
		stateInit,
		fsm.Events{
			{Name: eventStart, Src: []string{stateInit, stateStart, stateSendConfirm}, Dst: stateStart},
			{Name: eventRegister, Src: []string{stateStart}, Dst: stateStart},
			{Name: eventSuggest, Src: []string{stateStart}, Dst: stateSuggest},
			{Name: eventSuggested, Src: []string{stateSuggest}, Dst: stateStart},
			{Name: eventSend, Src: []string{stateStart}, Dst: stateSendTo},
			{Name: eventSendTo, Src: []string{stateSendTo}, Dst: stateSendAmount},
			{Name: eventSendAmount, Src: []string{stateSendAmount}, Dst: stateSendConfirm},
			{Name: eventSendConfirm, Src: []string{stateSendConfirm}, Dst: stateStart},
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
					UserID:   st.UserID,
					Username: st.Data["username"].(string),
					Address:  pair.Address(),
					Seed:     pair.Seed(),
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
			eventSend: func(ctx context.Context, e *fsm.Event) {
				st := e.Args[0].(db.State)
				msg := tgbotapi.NewMessage(st.UserID, textSend)
				msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
				if _, err := t.bot.Send(msg); err != nil {
					e.Cancel(err)
					return
				}
			},
			eventSendTo: func(ctx context.Context, e *fsm.Event) {
				spew.Dump(e.Args)

				st := e.Args[0].(db.State)
				key := e.Args[1].(string)

				if _, err := t.q.GetAccountByKey(ctx, key); err != nil {
					msg := tgbotapi.NewMessage(st.UserID, textSendToNotFount)
					msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
					if _, err := t.bot.Send(msg); err != nil {
						e.Cancel(err)
						return
					}
					e.Cancel(err)
					return
				}

				st.Data["send_to"] = key

				if err := t.q.UpdateStateData(ctx, db.UpdateStateDataParams{
					Data:   st.Data,
					UserID: st.UserID,
				}); err != nil {
					e.Cancel(err)
					return
				}

				msg := tgbotapi.NewMessage(st.UserID, textSendAmount)
				msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
				if _, err := t.bot.Send(msg); err != nil {
					e.Cancel(err)
					return
				}
			},
			eventSendAmount: func(ctx context.Context, e *fsm.Event) {
				st := e.Args[0].(db.State)
				amount, ok := e.Args[1].(string)
				if !ok {
					e.Cancel(errors.New("invalid amount"))
					return
				}

				st.Data["send_amount"] = amount

				// TODO: check if amount is valid

				if err := t.q.UpdateStateData(ctx, db.UpdateStateDataParams{
					Data:   st.Data,
					UserID: st.UserID,
				}); err != nil {
					e.Cancel(err)
					return
				}

				msg := tgbotapi.NewMessage(st.UserID, fmt.Sprintf(textSendSummary, st.Data["send_to"], amount))
				msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
				if _, err := t.bot.Send(msg); err != nil {
					e.Cancel(err)
					return
				}
			},
			eventSendConfirm: func(ctx context.Context, e *fsm.Event) {
				st := e.Args[0].(db.State)

				acc, err := t.q.GetAccount(ctx, st.UserID)
				if err != nil {
					e.Cancel(err)
					return
				}

				if err := t.stellar.Send(ctx, acc.Seed, st.Data["send_to"].(string), st.Data["send_amount"].(string)); err != nil {
					e.Cancel(err)
					return
				}

				delete(st.Data, "send_to")
				delete(st.Data, "send_amount")

				if err := t.q.UpdateStateData(ctx, db.UpdateStateDataParams{
					Data:   st.Data,
					UserID: st.UserID,
				}); err != nil {
					e.Cancel(err)
					return
				}

				msg := tgbotapi.NewMessage(st.UserID, textSendSuccess)
				msg.ReplyMarkup = buttonsStart
				if _, err := t.bot.Send(msg); err != nil {
					e.Cancel(err)
					return
				}
			},
		},
	)
}
