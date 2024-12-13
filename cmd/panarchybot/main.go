package main

import (
	"context"
	"database/sql"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"strings"

	"github.com/go-telegram/bot"
	"github.com/jackc/pgx/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/pressly/goose/v3"
	"github.com/samber/lo"
	"github.com/stellar/go/clients/horizonclient"
	"github.com/xdefrag/panarchybot"
	"github.com/xdefrag/panarchybot/campaign"
	"github.com/xdefrag/panarchybot/chatgpt"
	"github.com/xdefrag/panarchybot/config"
	"github.com/xdefrag/panarchybot/db"
	"github.com/xdefrag/panarchybot/stellar"
	"github.com/xdefrag/panarchybot/tgbot"
)

var Commit string

func main() {
	configPathPtr := flag.String("config", "", "path to config file")
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	l := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelDebug,
	})).With("commit", Commit)

	cfg, err := config.Get(lo.FromPtr(configPathPtr))
	if err != nil {
		l.ErrorContext(ctx, err.Error())
		os.Exit(1)
	}

	_ = godotenv.Load()

	goose.SetDialect("pgx")
	goose.SetBaseFS(panarchybot.EmbedMigrations)

	conn, err := sql.Open("pgx", os.Getenv("POSTGRES_DSN"))
	if err != nil {
		l.ErrorContext(ctx, err.Error())
		os.Exit(1)
	}

	if err := goose.UpContext(ctx, conn, "migrations"); err != nil { //
		l.ErrorContext(ctx, err.Error())
		os.Exit(1)
	}
	if err := goose.VersionContext(ctx, conn, "migrations"); err != nil {
		l.ErrorContext(ctx, err.Error())
		os.Exit(1)
	}

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
