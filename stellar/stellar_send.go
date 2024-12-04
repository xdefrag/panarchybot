package stellar

import (
	"context"

	"github.com/stellar/go/clients/horizonclient"
	"github.com/stellar/go/keypair"
	"github.com/stellar/go/txnbuild"
)

func (c *Stellar) Send(ctx context.Context, fromSeed, toAddress, amount string) error {
	fromPair, err := keypair.ParseFull(fromSeed)
	if err != nil {
		return err
	}

	asset := &txnbuild.CreditAsset{
		Code:   c.cfg.Stellar.FundAccount.AssetCode,
		Issuer: c.cfg.Stellar.FundAccount.AssetIssuer,
	}

	fromAccountDetails, err := c.cl.AccountDetail(horizonclient.AccountRequest{
		AccountID: fromPair.Address(),
	})
	if err != nil {
		return err
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
		Memo:    txnbuild.MemoText("live and let live"),
	})
	if err != nil {
		return err
	}

	tx, err = tx.Sign(c.cfg.Stellar.FundAccount.Passphrase, fromPair)
	if err != nil {
		return err
	}

	_, err = c.cl.SubmitTransaction(tx)
	if err != nil {
		return err
	}

	return nil
}
