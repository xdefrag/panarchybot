package stellar_test

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/stellar/go/clients/horizonclient"
	"github.com/stellar/go/keypair"
	"github.com/stellar/go/network"
	"github.com/stellar/go/txnbuild"
	"github.com/stretchr/testify/require"
	"github.com/xdefrag/panarchybot/config"
	"github.com/xdefrag/panarchybot/stellar"
)

func TestSend(t *testing.T) {
	t.SkipNow()

	ctx := context.Background()
	cl := horizonclient.DefaultTestNetClient

	from := keypair.MustRandom()
	_, err := cl.Fund(from.Address())
	require.NoError(t, err)

	to := keypair.MustRandom()
	_, err = cl.Fund(to.Address())
	require.NoError(t, err)

	asset := txnbuild.CreditAsset{
		Code:   "PANARCHY",
		Issuer: from.Address(),
	}

	submitTransaction(t, to, []txnbuild.Operation{
		&txnbuild.ChangeTrust{
			Line:          txnbuild.ChangeTrustAssetWrapper{Asset: asset},
			SourceAccount: to.Address(),
		},
	})

	cfg := &config.Config{}
	cfg.Stellar.FundAccount.AssetCode = "PANARCHY"
	cfg.Stellar.FundAccount.AssetIssuer = from.Address()
	cfg.Stellar.FundAccount.BaseFee = 1000
	cfg.Stellar.FundAccount.Passphrase = network.TestNetworkPassphrase

	l := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))

	s := stellar.New(cl, cfg, l)
	hash, err := s.Send(ctx, from.Seed(), to.Address(), "1")
	require.NoError(t, err)
	require.NotEmpty(t, hash)
}

func submitTransaction(t *testing.T, signer *keypair.Full, ops []txnbuild.Operation) {
	t.Helper()

	cl := horizonclient.DefaultTestNetClient

	signerAccountDetauls, err := cl.AccountDetail(horizonclient.AccountRequest{
		AccountID: signer.Address(),
	})
	require.NoError(t, err)

	tx, err := txnbuild.NewTransaction(txnbuild.TransactionParams{
		SourceAccount:        &signerAccountDetauls,
		IncrementSequenceNum: true,
		Operations:           ops,
		Preconditions: txnbuild.Preconditions{
			TimeBounds: txnbuild.NewInfiniteTimeout(),
		},
		BaseFee: 1000,
	})
	require.NoError(t, err)

	tx, err = tx.Sign(network.TestNetworkPassphrase, signer)
	require.NoError(t, err)

	_, err = cl.SubmitTransaction(tx)
	require.NoError(t, err)
}
