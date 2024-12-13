package stellar

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/stellar/go/clients/horizonclient"
	"github.com/stellar/go/keypair"
	"github.com/stellar/go/txnbuild"
)

func (s *Stellar) CreateAccount(ctx context.Context) (*keypair.Full, error) {
	l := s.l.WithGroup("stellar").With(
		slog.String("distributor", s.cfg.Stellar.FundAccount.Address),
		slog.String("asset_code", s.cfg.Stellar.FundAccount.AssetCode),
		slog.String("asset_issuer", s.cfg.Stellar.FundAccount.AssetIssuer),
	)

	pair, err := s.createAccount(ctx)
	if err != nil {
		details := &strings.Builder{}
		if p := horizonclient.GetError(err); p != nil {
			fmt.Fprintf(details, "status: %d, type: %s, title: %s, detail: %s",
				p.Problem.Status, p.Problem.Type, p.Problem.Title, p.Problem.Detail)
		}
		l.ErrorContext(ctx, "failed to submit transaction",
			slog.String("error", err.Error()),
			"details", details.String())
		return nil, err
	}
	l.InfoContext(ctx, "transaction submitted")

	return pair, nil
}

func (s *Stellar) createAccount(_ context.Context) (*keypair.Full, error) {
	pair := keypair.MustRandom()
	pairMain, err := keypair.ParseFull(s.cfg.Stellar.FundAccount.Seed)
	if err != nil {
		return nil, err
	}

	mainAccountDetails, err := s.cl.AccountDetail(horizonclient.AccountRequest{
		AccountID: pairMain.Address(),
	})
	if err != nil {
		return nil, err
	}

	asset := txnbuild.CreditAsset{
		Code:   s.cfg.Stellar.FundAccount.AssetCode,
		Issuer: s.cfg.Stellar.FundAccount.AssetIssuer,
	}

	tx, err := txnbuild.NewTransaction(txnbuild.TransactionParams{
		SourceAccount:        &mainAccountDetails,
		IncrementSequenceNum: true,
		Operations: []txnbuild.Operation{
			&txnbuild.CreateAccount{
				Destination:   pair.Address(),
				Amount:        s.cfg.Stellar.FundAccount.DefaultAmount,
				SourceAccount: pairMain.Address(),
			},
			&txnbuild.ChangeTrust{
				Line:          txnbuild.ChangeTrustAssetWrapper{Asset: asset},
				SourceAccount: pair.Address(),
			},
			&txnbuild.Payment{
				Destination:   pair.Address(),
				Amount:        s.cfg.Stellar.FundAccount.DefaultAirdrop,
				Asset:         asset,
				SourceAccount: pairMain.Address(),
			},
		},
		Preconditions: txnbuild.Preconditions{
			TimeBounds: txnbuild.NewInfiniteTimeout(),
		},
		BaseFee: s.cfg.Stellar.FundAccount.BaseFee,
		Memo:    txnbuild.MemoText(s.cfg.Stellar.FundAccount.Memo),
	})
	if err != nil {
		return nil, err
	}

	tx, err = tx.Sign(s.cfg.Stellar.FundAccount.Passphrase, pairMain, pair)
	if err != nil {
		return nil, err
	}

	_, err = s.cl.SubmitTransaction(tx)
	if err != nil {
		return nil, err
	}

	return pair, nil
}
