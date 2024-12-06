package campaign

import (
	"context"
	"fmt"
	"strconv"

	"github.com/xdefrag/panarchybot/config"
)

type Campaign struct {
	cfg *config.Config
}

type AirdropParams struct {
	Username string
	UserID   int64
	ID       int64
}

type AirdropResult struct {
	Amount string
	Text   string
	Memo   string
}

func (c *Campaign) Airdrop(ctx context.Context, p AirdropParams) (AirdropResult, error) {
	res := AirdropResult{}

	if c.cfg.Stellar.FundAccount.Airdrop.Enable {
		return res, nil
	}

	amount, ok := c.cfg.Stellar.FundAccount.Airdrop.ByUsernameAmount[p.Username]
	if ok {
		res.Amount = amount
		res.Text = fmt.Sprintf("Тебе назначен персональный эйрдроп %s %s. Спасибо за поддержку! ❤️",
			amount, c.cfg.Stellar.FundAccount.AssetCode)
		res.Memo = "best human alive airdrop"
		return res, nil
	}

	for idLessStr, amount := range c.cfg.Stellar.FundAccount.Airdrop.IDLessThanAmount {
		idLess, err := strconv.ParseInt(idLessStr, 10, 64)
		if err != nil {
			return res, err
		}

		if p.ID < idLess {
			res.Amount = amount
			res.Text = fmt.Sprintf("Ты зарегистрировался в первой %d-ке! Поздравляю, слоняра, держи %s %s.",
				idLess, amount, c.cfg.Stellar.FundAccount.AssetCode)
			res.Memo = fmt.Sprintf("first %d airdrop", idLess)
			break
		}
	}

	return res, nil
}

func New(
	cfg *config.Config,
) *Campaign {
	return &Campaign{
		cfg: cfg,
	}
}
