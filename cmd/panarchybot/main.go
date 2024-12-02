package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/xdefrag/panarchybot/chatgpt"
	"github.com/xdefrag/panarchybot/tgbot"
)

var Commit string

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	l := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})).With("commit", Commit)

	if err := godotenv.Load(); err != nil {
		l.ErrorContext(ctx, err.Error())
		os.Exit(1)
	}

	gpt := chatgpt.New(openai.NewClient(
		option.WithAPIKey(os.Getenv("OPENAI_API_KEY")),
	))

	bot, err := tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_TOKEN"))
	if err != nil {
		l.ErrorContext(ctx, err.Error())
		os.Exit(1)
	}

	tgbot := tgbot.New(l, nil, bot, gpt)

	tgbot.Run(ctx) // blocking
}
