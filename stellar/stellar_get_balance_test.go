package stellar_test

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/stellar/go/clients/horizonclient"
	"github.com/stellar/go/keypair"
	"github.com/stretchr/testify/require"
	"github.com/xdefrag/panarchybot/config"
	"github.com/xdefrag/panarchybot/stellar"
)

func TestGetBalance(t *testing.T) {
	t.SkipNow()

	ctx := context.Background()
	cl := horizonclient.DefaultTestNetClient

	pair := keypair.MustRandom()
	_, err := cl.Fund(pair.Address())
	require.NoError(t, err)

	cfg := &config.Config{}
	cfg.Stellar.FundAccount.AssetCode = "TEST"
	cfg.Stellar.FundAccount.AssetIssuer = pair.Address()

	l := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))

	s := stellar.New(cl, cfg, l)
	res, err := s.GetBalance(ctx, pair.Address())
	require.NoError(t, err)
	require.NotNil(t, res)
}
