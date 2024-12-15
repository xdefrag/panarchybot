package tgbot

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/samber/lo"
	"github.com/xdefrag/panarchybot/db"
)

const (
	stateStart       = "start"
	stateSuggest     = "suggest"
	stateSendTo      = "send_to"
	stateSendAmount  = "send_amount"
	stateSendConfirm = "send_confirm"
)

func (t *TGBot) callbackSuggestPrivateHandler(ctx context.Context, st db.State, upd *models.Update) error {
	if _, err := t.bot.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: upd.CallbackQuery.ID,
	}); err != nil {
		return err
	}

	if err := t.q.CreateState(ctx, db.CreateStateParams{
		UserID: st.UserID,
		State:  stateSuggest,
		Data:   st.Data,
		Meta:   st.Meta,
	}); err != nil {
		return err
	}

	_, err := t.bot.EditMessageText(ctx, &bot.EditMessageTextParams{
		ChatID:      upd.CallbackQuery.From.ID,
		MessageID:   upd.CallbackQuery.Message.Message.ID,
		ReplyMarkup: nil,
		Text:        textSuggestWelcome,
	})

	return err
}

func (t *TGBot) callbackSendPrivateHandler(ctx context.Context, st db.State, upd *models.Update) error {
	if _, err := t.bot.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: upd.CallbackQuery.ID,
	}); err != nil {
		return err
	}

	if err := t.q.CreateState(ctx, db.CreateStateParams{
		UserID: st.UserID,
		State:  stateSendTo,
		Data:   st.Data,
		Meta:   st.Meta,
	}); err != nil {
		return err
	}

	_, err := t.bot.EditMessageText(ctx, &bot.EditMessageTextParams{
		ChatID:      upd.CallbackQuery.From.ID,
		MessageID:   upd.CallbackQuery.Message.Message.ID,
		ReplyMarkup: nil,
		Text:        textSendToWhom,
	})

	return err
}

const stellarExpertTxTemplate = "https://stellar.expert/explorer/%s/search?term=%s"

func (t *TGBot) callbackSendConfirmPrivateHandler(ctx context.Context, st db.State, upd *models.Update) error {
	if _, err := t.bot.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: upd.CallbackQuery.ID,
	}); err != nil {
		return err
	}

	if _, err := t.bot.EditMessageText(ctx, &bot.EditMessageTextParams{
		ChatID:      upd.CallbackQuery.From.ID,
		MessageID:   upd.CallbackQuery.Message.Message.ID,
		Text:        textSendNotifySendingTx,
		ReplyMarkup: nil,
	}); err != nil {
		return err
	}

	accFrom, err := t.q.GetAccount(ctx, st.UserID)
	if err != nil {
		return err
	}

	sendTo := st.Data["send_to"].(string)
	sendAmount := st.Data["send_amount"].(string)

	hash, err := t.stellar.Send(ctx, accFrom.Seed, sendTo, sendAmount)
	if err != nil {
		if _, err := t.bot.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: upd.CallbackQuery.From.ID,
			Text:   textSendError,
		}); err != nil {
			return err
		}

		return t.startPrivateHandler(ctx, st, upd)
	}

	if _, err := t.bot.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: upd.CallbackQuery.From.ID,
		Text: fmt.Sprintf(textTemplateSendSuccess,
			fmt.Sprintf(stellarExpertTxTemplate, t.cfg.Stellar.FundAccount.Network, hash),
		),
		ParseMode:          models.ParseModeHTML,
		LinkPreviewOptions: &models.LinkPreviewOptions{IsDisabled: lo.ToPtr(true)},
	}); err != nil {
		return err
	}

	accTo, err := t.q.GetAccountByKey(ctx, sendTo)
	if err != nil {
		t.l.ErrorContext(ctx, "failed to get account by key",
			slog.String("error", err.Error()))
		return t.startPrivateHandler(ctx, st, upd)
	}

	if _, err := t.bot.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: accTo.UserID,
		Text: fmt.Sprintf(textTemplateThanksReceived,
			fmt.Sprintf(stellarExpertTxTemplate, t.cfg.Stellar.FundAccount.Network, hash),
			sendAmount, t.cfg.Stellar.FundAccount.AssetCode, accFrom.Username),
		ParseMode:          models.ParseModeHTML,
		LinkPreviewOptions: &models.LinkPreviewOptions{IsDisabled: lo.ToPtr(true)},
	}); err != nil {
		t.l.ErrorContext(ctx, "failed to notify to-user about thanks transaction",
			slog.String("error", err.Error()))
	}

	return t.startPrivateHandler(ctx, st, upd)
}
