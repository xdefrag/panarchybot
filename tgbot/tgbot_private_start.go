package tgbot

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/jackc/pgx/v5"
	"github.com/samber/lo"
	"github.com/xdefrag/panarchybot/campaign"
	"github.com/xdefrag/panarchybot/db"
	"github.com/xdefrag/panarchybot/stellar"
)

var startKeyboard = &models.InlineKeyboardMarkup{
	InlineKeyboard: [][]models.InlineKeyboardButton{
		{
			{Text: textStartSend, CallbackData: "send"},
			{Text: textStartSuggest, CallbackData: "suggest"},
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
			{Text: textStartRegister, CallbackData: "register"},
		},
	},
}

func (t *TGBot) startRegister(ctx context.Context, st db.State) error {
	msg, err := t.bot.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      st.UserID,
		Text:        fmt.Sprintf(textStartWelcome, t.cfg.Stellar.FundAccount.AssetCode),
		ReplyMarkup: registerKeyboard,
		ParseMode:   models.ParseModeHTML,
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

	fmt.Fprintf(text, textStartDashboard, stellarExpertURLPrefix, acc.Address,
		addrAbbr(acc.Address), bal, t.cfg.Stellar.FundAccount.AssetCode)

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

	id, err := t.q.CreateAccount(ctx, db.CreateAccountParams{
		UserID:   st.UserID,
		Username: st.Data["username"].(string),
		Address:  pair.Address(),
		Seed:     pair.Seed(),
	})

	airdrop, err := t.campaign.Airdrop(ctx, campaign.AirdropParams{
		Username: st.Data["username"].(string),
		UserID:   st.UserID,
		ID:       id,
	})
	if err != nil {
		return err
	}

	if airdrop.Amount != "" {
		_, err = t.stellar.Send(
			ctx,
			t.cfg.Stellar.FundAccount.Seed,
			pair.Address(),
			airdrop.Amount,
			stellar.WithMemo(airdrop.Memo),
		)
		if err != nil {
			t.l.ErrorContext(ctx, "failed to send airdrop",
				slog.String("error", err.Error()),
				slog.Int64("user_id", st.UserID),
				slog.String("username", st.Data["username"].(string)),
				slog.Int64("update_id", upd.ID))
		}

		if err == nil {
			_, err := t.bot.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: st.UserID,
				Text:   airdrop.Text,
			})
			if err != nil {
				return err
			}
		}
	}

	return t.startPrivateHandler(ctx, st, upd)
}

func addrAbbr(addr string) string {
	return fmt.Sprintf("%s...%s", addr[:4], addr[len(addr)-4:])
}
