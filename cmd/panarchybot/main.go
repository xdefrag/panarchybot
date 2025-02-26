package main

import (
	"context"
	"database/sql"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/go-telegram/bot"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/pressly/goose/v3"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/samber/lo"
	"github.com/stellar/go/clients/horizonclient"
	"github.com/xdefrag/panarchybot"
	"github.com/xdefrag/panarchybot/campaign"
	"github.com/xdefrag/panarchybot/chatgpt"
	"github.com/xdefrag/panarchybot/config"
	"github.com/xdefrag/panarchybot/db"
	"github.com/xdefrag/panarchybot/metrics"
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

	_ = goose.SetDialect("pgx")
	goose.SetBaseFS(panarchybot.EmbedMigrations)

	conn, err := sql.Open("pgx", os.Getenv("POSTGRES_DSN"))
	if err != nil {
		l.ErrorContext(ctx, err.Error())
		os.Exit(1)
	}
	defer conn.Close()

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

	poolConfig, err := pgxpool.ParseConfig(os.Getenv("POSTGRES_DSN"))
	if err != nil {
		l.ErrorContext(ctx, err.Error())
		os.Exit(1)
	}

	// Настраиваем пул
	poolConfig.MaxConns = 10
	poolConfig.MinConns = 2
	poolConfig.MaxConnLifetime = time.Hour
	poolConfig.MaxConnIdleTime = time.Minute * 30
	poolConfig.HealthCheckPeriod = time.Second * 30
	poolConfig.ConnConfig.ConnectTimeout = time.Second * 5

	pg, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		l.ErrorContext(ctx, err.Error())
		os.Exit(1)
	}
	defer pg.Close()

	// Запускаем сбор метрик пула
	metrics.StartPoolMetrics(ctx, pg)

	horizonClient := horizonclient.DefaultPublicNetClient
	if strings.Contains(cfg.Stellar.FundAccount.Passphrase, "Test") {
		horizonClient = horizonclient.DefaultTestNetClient
	}

	ledger := stellar.New(horizonClient, cfg, l)
	ledgerWithMetrics := metrics.NewLedgerWrapper(ledger)

	camp := campaign.New(cfg)

	queries := db.New(pg)
	queriesWithTimeout := db.WithTimeout(queries, time.Second*15)

	tgbot := tgbot.New(cfg, queriesWithTimeout, bot, ledgerWithMetrics, gpt, camp, l)

	go func() {
		mux := http.NewServeMux()
		mux.Handle("/metrics", promhttp.Handler())
		_ = http.ListenAndServe(cfg.Metrics.Addr, mux)
	}()

	tgbot.Run(ctx) // blocks
}
