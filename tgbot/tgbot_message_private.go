package tgbot

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/samber/lo"
	"github.com/xdefrag/panarchybot/db"
)

var ErrNotExpectingInput = errors.New("not expecting input")

func (t *TGBot) messagePrivateHandler(ctx context.Context, st db.State, upd *models.Update, l *slog.Logger) error {
	switch st.State {
	case stateSuggest:
		return t.suggestedPrivateHandler(ctx, st, upd, l)
	case stateSendTo:
		return t.sendToPrivateHandler(ctx, st, upd, l)
	case stateSendAmount:
		return t.sendAmountPrivateHandler(ctx, st, upd, l)
	default:
		return ErrNotExpectingInput
	}
}

func (t *TGBot) suggestedPrivateHandler(ctx context.Context, st db.State, upd *models.Update, l *slog.Logger) error {
	if _, err := t.bot.ForwardMessage(ctx, &bot.ForwardMessageParams{
		ChatID:     t.cfg.Telegram.SuggestChatID,
		FromChatID: st.UserID,
		MessageID:  upd.Message.ID,
	}); err != nil {
		return err
	}

	if _, err := t.bot.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: st.UserID,
		Text:   textSuggestSubmited,
	}); err != nil {
		return err
	}

	return t.startPrivateHandler(ctx, st, upd, l)
}

func (t *TGBot) sendToPrivateHandler(ctx context.Context, st db.State, upd *models.Update, l *slog.Logger) error {
	key := strings.ReplaceAll(upd.Message.Text, "@", "")

	acc, err := t.q.GetAccountByKey(ctx, key)
	if err != nil {
		if _, err := t.bot.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: st.UserID,
			Text:   textSendErrorUserNotFound,
		}); err != nil {
			return err
		}

		return nil
	}

	st.Data["send_to"] = acc.Address

	if err := t.q.CreateState(ctx, db.CreateStateParams{
		UserID: st.UserID,
		State:  stateSendAmount,
		Data:   st.Data,
		Meta:   st.Meta,
	}); err != nil {
		return err
	}

	if _, err := t.bot.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: st.UserID,
		Text:   textSendAmount,
	}); err != nil {
		return err
	}

	return nil
}

var sendConfirmKeyboard = &models.InlineKeyboardMarkup{
	InlineKeyboard: [][]models.InlineKeyboardButton{
		{
			{Text: textSendYes, CallbackData: "send_confirm"},
			{Text: textSendNo, CallbackData: "start"},
		},
	},
}

func (t *TGBot) sendAmountPrivateHandler(ctx context.Context, st db.State, upd *models.Update, l *slog.Logger) error {
	acc, err := t.q.GetAccount(ctx, st.UserID)
	if err != nil {
		return err
	}

	amountStr := upd.Message.Text

	balStr, err := t.ledger.GetBalance(ctx, acc.Address)
	if err != nil {
		return err
	}

	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		return err
	}

	bal, err := strconv.ParseFloat(balStr, 64)
	if err != nil {
		return err
	}

	if amount > bal {
		if _, err := t.bot.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: st.UserID,
			Text:   textSendErrorNotEnough,
		}); err != nil {
			return err
		}

		return nil
	}

	st.Data["send_amount"] = amountStr

	msg, err := t.bot.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: st.UserID,
		Text: fmt.Sprintf(textTemplateSendConfirm,
			stellarExpertURLPrefix, st.Data["send_to"],
			addrAbbr(st.Data["send_to"].(string)),
			st.Data["send_amount"],
			t.cfg.Stellar.FundAccount.AssetCode,
		),
		ParseMode:          models.ParseModeHTML,
		ReplyMarkup:        sendConfirmKeyboard,
		LinkPreviewOptions: &models.LinkPreviewOptions{IsDisabled: lo.ToPtr(true)},
	})
	if err != nil {
		return err
	}

	st.Data["menu_message_id"] = msg.ID

	if err := t.q.CreateState(ctx, db.CreateStateParams{
		UserID: st.UserID,
		State:  stateSendConfirm,
		Data:   st.Data,
		Meta:   st.Meta,
	}); err != nil {
		return err
	}

	return nil
}
