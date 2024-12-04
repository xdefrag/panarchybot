package tgbot

import (
	"context"
	"fmt"

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
		Text:        "Предложи новость или задай вопрос",
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
		Text:        "Кому отправить?",
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
		Text:        "Отправляем транзакцию",
		ReplyMarkup: nil,
	}); err != nil {
		return err
	}

	acc, err := t.q.GetAccount(ctx, st.UserID)
	if err != nil {
		return err
	}

	hash, err := t.stellar.Send(ctx, acc.Seed, st.Data["send_to"].(string), st.Data["send_amount"].(string))
	if err != nil {
		if _, err := t.bot.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: upd.CallbackQuery.From.ID,
			Text:   "Произошла ошибка при отправке",
		}); err != nil {
			return err
		}

		return t.startPrivateHandler(ctx, st, upd)
	}

	if _, err := t.bot.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: upd.CallbackQuery.From.ID,
		Text: fmt.Sprintf("<a href=\"%s\">Транзакция</a> успешно отправлена",
			fmt.Sprintf(stellarExpertTxTemplate, t.cfg.Stellar.FundAccount.Network, hash),
		),
		ParseMode:          models.ParseModeHTML,
		LinkPreviewOptions: &models.LinkPreviewOptions{IsDisabled: lo.ToPtr(true)},
	}); err != nil {
		return err
	}

	return t.startPrivateHandler(ctx, st, upd)
}
