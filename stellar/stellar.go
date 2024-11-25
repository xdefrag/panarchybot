package stellar

import (
	"context"
	"log/slog"

	"github.com/stellar/go/clients/horizonclient"
	"github.com/stellar/go/keypair"
	"github.com/stellar/go/txnbuild"
	"github.com/xdefrag/panarchybot/config"
)

type Stellar struct {
	cl  horizonclient.ClientInterface
	cfg *config.Config
	l   *slog.Logger
}

func (s *Stellar) CreateAccount(ctx context.Context, amount string) (*keypair.Full, error) {
	pair := keypair.MustRandom()
	pairMain, err := keypair.ParseFull(s.cfg.Stellar.FundAccount.Seed)
	if err != nil {
		return nil, err
	}

	asset := txnbuild.CreditAsset{
		Code:   s.cfg.Stellar.FundAccount.AssetCode,
		Issuer: s.cfg.Stellar.FundAccount.AssetIssuer,
	}

	l := s.l.WithGroup("stellar").With(
		slog.String("account", pair.Address()),
		slog.String("distributor", pairMain.Address()),
		slog.String("asset_code", s.cfg.Stellar.FundAccount.AssetCode),
		slog.String("asset_issuer", s.cfg.Stellar.FundAccount.AssetIssuer),
	)

	mainAccountDetails, err := s.cl.AccountDetail(horizonclient.AccountRequest{
		AccountID: pairMain.Address(),
	})
	if err != nil {
		return nil, err
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
				Amount:        amount,
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
		l.ErrorContext(ctx, "failed to submit transaction", slog.String("error", err.Error()))
		return nil, err
	}
	l.InfoContext(ctx, "transaction submitted")

	return pair, nil
}

func New(cl horizonclient.ClientInterface, cfg *config.Config, l *slog.Logger) *Stellar {
	return &Stellar{
		cl:  cl,
		cfg: cfg,
		l:   l,
	}
}
