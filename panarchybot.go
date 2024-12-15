package panarchybot

import (
	"context"
	"embed"
	"log/slog"

	"github.com/go-telegram/bot/models"
	"github.com/stellar/go/keypair"
	"github.com/xdefrag/panarchybot/db"
)

//go:embed migrations/*.sql
var EmbedMigrations embed.FS

type SendOptions struct {
	Memo string
}

type SendOption func(*SendOptions)

func WithMemo(memo string) SendOption {
	return func(o *SendOptions) {
		o.Memo = memo
	}
}

type Ledger interface {
	CreateAccount(ctx context.Context) (*keypair.Full, error)
	GetBalance(ctx context.Context, address string) (string, error)
	Send(ctx context.Context, fromSeed, toAddress, amount string, opts ...SendOption) (string, error)
}

type (
	TelegramBotPrivateHandler    func(ctx context.Context, state db.State, upd *models.Update, l *slog.Logger) error
	TelegramBotGroupHandler      func(ctx context.Context, upd *models.Update, l *slog.Logger) error
	TelegramBotPrivateMiddleware func(next TelegramBotPrivateHandler) TelegramBotPrivateHandler
	TelegramBotGroupMiddleware   func(next TelegramBotGroupHandler) TelegramBotGroupHandler
)
