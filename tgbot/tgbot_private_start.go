package tgbot

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/jackc/pgx/v5"
	"github.com/samber/lo"
	"github.com/xdefrag/panarchybot/db"
)

var startKeyboard = &models.InlineKeyboardMarkup{
	InlineKeyboard: [][]models.InlineKeyboardButton{
		{
			{Text: "Отправить", CallbackData: "send"},
			{Text: "Предложка", CallbackData: "suggest"},
		},
	},
}

const stellarExpertURLPrefix = "https://stellar.expert/explorer/public/account/"

func (t *TGBot) startPrivateHandler(ctx context.Context, st db.State, upd *models.Update) error {
	user := getUser(upd)

	if st, err := t.q.GetState(ctx, user.ID); err == nil {
		menuID, ok := st.Data["menu_message_id"]
		if ok {
			if _, err := t.bot.DeleteMessage(ctx, &bot.DeleteMessageParams{
				ChatID:    user.ID,
				MessageID: int(menuID.(float64)),
			}); err != nil {
				return err
			}
		}
	}

	data := make(map[string]interface{})
	meta := make(map[string]interface{})

	data["username"] = user.Username
	meta["firstname"] = user.FirstName
	meta["lastname"] = user.LastName

	if err := t.q.CreateState(ctx, db.CreateStateParams{
		UserID: user.ID,
		State:  stateStart,
		Data:   data,
		Meta:   meta,
	}); err != nil {
		return err
	}

	acc, err := t.q.GetAccount(ctx, user.ID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return err
	}

	if errors.Is(err, pgx.ErrNoRows) {
		return t.startRegister(ctx, st)
	}

	return t.startBalance(ctx, acc, data)
}

var registerKeyboard = &models.InlineKeyboardMarkup{
	InlineKeyboard: [][]models.InlineKeyboardButton{
		{
			{Text: "Регистрация", CallbackData: "register"},
		},
	},
}

func (t *TGBot) startRegister(ctx context.Context, st db.State) error {
	msg, err := t.bot.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      st.UserID,
		Text:        "Зарегистрируйся",
		ReplyMarkup: registerKeyboard,
	})
	if err != nil {
		return err
	}

	st.Data["menu_message_id"] = msg.ID

	if err := t.q.UpdateStateData(ctx, db.UpdateStateDataParams{
		UserID: st.UserID,
		Data:   st.Data,
	}); err != nil {
		return err
	}

	return nil
}

func (t *TGBot) startBalance(ctx context.Context, acc db.Account, data map[string]interface{}) error {
	bal, err := t.stellar.GetBalance(ctx, acc.Address)
	if err != nil {
		return err
	}

	text := &strings.Builder{}

	fmt.Fprintf(text, "Счет <a href=\"%s%s\">%s</a>\n", stellarExpertURLPrefix, acc.Address, addrAbbr(acc.Address))
	fmt.Fprintf(text, "%s PANARCHY", bal)

	msg, err := t.bot.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:             acc.UserID,
		Text:               text.String(),
		ReplyMarkup:        startKeyboard,
		ParseMode:          models.ParseModeHTML,
		LinkPreviewOptions: &models.LinkPreviewOptions{IsDisabled: lo.ToPtr(true)},
	})
	if err != nil {
		return err
	}

	data["menu_message_id"] = msg.ID

	if err := t.q.UpdateStateData(ctx, db.UpdateStateDataParams{
		UserID: acc.UserID,
		Data:   data,
	}); err != nil {
		return err
	}

	return nil
}

func (t *TGBot) registerPrivateHandler(ctx context.Context, st db.State, upd *models.Update) error {
	if _, err := t.bot.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: upd.CallbackQuery.ID,
		Text:            "Регистрируем",
	}); err != nil {
		return err
	}

	if _, err := t.bot.EditMessageReplyMarkup(ctx, &bot.EditMessageReplyMarkupParams{
		ChatID:      upd.CallbackQuery.From.ID,
		MessageID:   upd.CallbackQuery.Message.Message.ID,
		ReplyMarkup: nil,
	}); err != nil {
		return err
	}

	pair, err := t.stellar.CreateAccount(ctx)
	if err != nil {
		return err
	}

	num, err := t.q.CreateAccount(ctx, db.CreateAccountParams{
		UserID:   st.UserID,
		Username: st.Data["username"].(string),
		Address:  pair.Address(),
		Seed:     pair.Seed(),
	})

	_ = num // TODO aurdrop

	return t.startPrivateHandler(ctx, st, upd)
}

func addrAbbr(addr string) string {
	return fmt.Sprintf("%s...%s", addr[:4], addr[len(addr)-4:])
}
