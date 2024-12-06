package stellar

import (
	"context"

	"github.com/stellar/go/clients/horizonclient"
	"github.com/stellar/go/keypair"
	"github.com/stellar/go/txnbuild"
)

type SendOptions struct {
	Memo string
}

type SendOption func(*SendOptions)

func WithMemo(memo string) SendOption {
	return func(o *SendOptions) {
		o.Memo = memo
	}
}

func (c *Stellar) Send(ctx context.Context, fromSeed, toAddress, amount string, opts ...SendOption) (string, error) {
	options := &SendOptions{
		Memo: "live and let live",
	}
	for _, o := range opts {
		o(options)
	}

	fromPair, err := keypair.ParseFull(fromSeed)
	if err != nil {
		return "", err
	}

	asset := &txnbuild.CreditAsset{
		Code:   c.cfg.Stellar.FundAccount.AssetCode,
		Issuer: c.cfg.Stellar.FundAccount.AssetIssuer,
	}

	fromAccountDetails, err := c.cl.AccountDetail(horizonclient.AccountRequest{
		AccountID: fromPair.Address(),
	})
	if err != nil {
		return "", err
	}

	tx, err := txnbuild.NewTransaction(txnbuild.TransactionParams{
		SourceAccount:        &fromAccountDetails,
		IncrementSequenceNum: true,
		Operations: []txnbuild.Operation{
			&txnbuild.Payment{
				Destination:   toAddress,
				Amount:        amount,
				Asset:         asset,
				SourceAccount: fromPair.Address(),
			},
		},
		Preconditions: txnbuild.Preconditions{
			TimeBounds: txnbuild.NewInfiniteTimeout(),
		},
		BaseFee: c.cfg.Stellar.FundAccount.BaseFee,
		Memo:    txnbuild.MemoText(options.Memo),
	})
	if err != nil {
		return "", err
	}

	tx, err = tx.Sign(c.cfg.Stellar.FundAccount.Passphrase, fromPair)
	if err != nil {
		return "", err
	}

	txr, err := c.cl.SubmitTransaction(tx)
	if err != nil {
		return "", err
	}

	return txr.Hash, nil
}
