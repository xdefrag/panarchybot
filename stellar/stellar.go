package stellar

import (
	"log/slog"

	"github.com/stellar/go/clients/horizonclient"
	"github.com/xdefrag/panarchybot/config"
)

type Stellar struct {
	cl  horizonclient.ClientInterface
	cfg *config.Config
	l   *slog.Logger
}

func New(cl horizonclient.ClientInterface, cfg *config.Config, l *slog.Logger) *Stellar {
	return &Stellar{
		cl:  cl,
		cfg: cfg,
		l:   l,
	}
}
