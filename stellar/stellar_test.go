package stellar_test

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/stellar/go/clients/horizonclient"
	"github.com/stellar/go/keypair"
	"github.com/stellar/go/network"
	"github.com/stretchr/testify/require"
	"github.com/xdefrag/panarchybot/config"
	"github.com/xdefrag/panarchybot/stellar"
)

func TestCreateAccount(t *testing.T) {
	ctx := context.Background()
	cl := horizonclient.DefaultTestNetClient

	pair := keypair.MustRandom()
	_, err := cl.Fund(pair.Address())
	require.NoError(t, err)

	cfg := &config.Config{}
	cfg.Stellar.FundAccount.Address = pair.Address()
	cfg.Stellar.FundAccount.Seed = pair.Seed()
	cfg.Stellar.FundAccount.AssetCode = "PANARCHY"
	cfg.Stellar.FundAccount.AssetIssuer = pair.Address()
	cfg.Stellar.FundAccount.BaseFee = 1000
	cfg.Stellar.FundAccount.DefaultAmount = "2"
	cfg.Stellar.FundAccount.Passphrase = network.TestNetworkPassphrase
	cfg.Stellar.FundAccount.Memo = "panarchynow.t.me"

	l := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))

	s := stellar.New(cl, cfg, l)
	res, err := s.CreateAccount(ctx, "1000")
	require.NoError(t, err)
	require.NotNil(t, res)

	spew.Dump(res)
}
