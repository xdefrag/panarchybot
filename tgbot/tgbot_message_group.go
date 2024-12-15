package tgbot

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/samber/lo"
	"github.com/stellar/go/clients/horizonclient"
)

const (
	thanksCmd    = "/thanks"
	thanksAmount = 1
)

func (t *TGBot) thanksGroupHandler(ctx context.Context, upd *models.Update, l *slog.Logger) error {
	if !t.cfg.Telegram.Thanks.Enable {
		return nil
	}

	if upd.Message.From == nil || upd.Message.ReplyToMessage.From == nil {
		return nil
	}

	amountF := t.getThanksAmount(upd.Message.Text)
	if amountF == 0 {
		return t.answerMessage(ctx, upd, textThanksErrorAmount)
	}

	amount := fmt.Sprintf("%f", amountF)

	from, err := t.q.GetAccount(ctx, upd.Message.From.ID)
	if err != nil {
		if err := t.answerMessage(ctx, upd, textThanksErrorAccountFrom); err != nil {
			return err
		}
		return err
	}

	to, err := t.q.GetAccount(ctx, upd.Message.ReplyToMessage.From.ID)
	if err != nil {
		if err := t.answerMessage(ctx, upd, textThanksErrorAccountTo); err != nil {
			return err
		}
		return err
	}

	hash, err := t.ledger.Send(ctx, from.Seed, to.Address, amount)
	if err != nil { // TODO: handle error
		errHor := horizonclient.GetError(err)
		l.ErrorContext(ctx, "failed to send stellar transaction",
			slog.String("error", errHor.Problem.Detail))
		return err
	}

	if _, err = t.bot.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: to.UserID,
		Text: fmt.Sprintf(textTemplateThanksReceived,
			fmt.Sprintf(stellarExpertTxTemplate, t.cfg.Stellar.FundAccount.Network, hash),
			amount, t.cfg.Stellar.FundAccount.AssetCode, upd.Message.From.Username),
		ParseMode:          models.ParseModeHTML,
		LinkPreviewOptions: &models.LinkPreviewOptions{IsDisabled: lo.ToPtr(true)},
	}); err != nil {
		l.ErrorContext(ctx, "failed to notify to-user about thanks transaction",
			slog.String("error", err.Error()))
	}

	return t.answerMessage(ctx, upd,
		fmt.Sprintf(textTemplateThanksSuccess,
			amount, t.cfg.Stellar.FundAccount.AssetCode, upd.Message.ReplyToMessage.From.Username))
}

func (t *TGBot) answerMessage(ctx context.Context, upd *models.Update, text string) error {
	_, err := t.bot.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: upd.Message.Chat.ID,
		ReplyParameters: &models.ReplyParameters{
			MessageID: upd.Message.ID,
		},
		Text: text,
	})

	return err
}

func (t *TGBot) getThanksAmount(cmd string) float64 {
	if cmd == thanksCmd {
		return thanksAmount
	}

	parts := strings.Split(cmd, " ")
	if len(parts) < 2 {
		return 0
	}

	amount, _ := strconv.ParseFloat(parts[1], 64)

	return amount
}
