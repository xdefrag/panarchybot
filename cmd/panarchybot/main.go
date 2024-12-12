package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"strings"

	"github.com/go-telegram/bot"
	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/stellar/go/clients/horizonclient"
	"github.com/xdefrag/panarchybot/campaign"
	"github.com/xdefrag/panarchybot/chatgpt"
	"github.com/xdefrag/panarchybot/config"
	"github.com/xdefrag/panarchybot/db"
	"github.com/xdefrag/panarchybot/stellar"
	"github.com/xdefrag/panarchybot/tgbot"
)

var Commit string

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	l := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})).With("commit", Commit)

	cfg, err := config.Get()
	if err != nil {
		l.ErrorContext(ctx, err.Error())
		os.Exit(1)
	}

	_ = godotenv.Load()

	if cfg.Stellar.FundAccount.Seed == "" {
		cfg.Stellar.FundAccount.Seed = os.Getenv("FUND_SEED")
	}

	gpt := chatgpt.New(openai.NewClient(
		option.WithAPIKey(os.Getenv("OPENAI_API_KEY")),
	), cfg)

	bot, err := bot.New(os.Getenv("TELEGRAM_TOKEN"))
	if err != nil {
		l.ErrorContext(ctx, err.Error())
		os.Exit(1)
	}

	pg, err := pgx.Connect(ctx, os.Getenv("POSTGRES_DSN"))
	if err != nil {
		l.ErrorContext(ctx, err.Error())
		os.Exit(1)
	}

	horizonClient := horizonclient.DefaultPublicNetClient
	if strings.Contains(cfg.Stellar.FundAccount.Passphrase, "Test") {
		horizonClient = horizonclient.DefaultTestNetClient
	}

	st := stellar.New(horizonClient, cfg, l)

	camp := campaign.New(cfg)

	tgbot := tgbot.New(cfg, db.New(pg), bot, st, gpt, camp, l)

	tgbot.Run(ctx) // blocks
}
