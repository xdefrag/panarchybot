package metrics

import (
	"context"
	"log/slog"

	"github.com/go-telegram/bot/models"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/xdefrag/panarchybot"
	"github.com/xdefrag/panarchybot/db"
)

var counterTGBotPrivateHandler = prometheus.NewCounterVec(prometheus.CounterOpts{
	Name: "panarchybot_tgbot_private_handler_total",
}, []string{"status"})

var counterTGBotGroupHandler = prometheus.NewCounterVec(prometheus.CounterOpts{
	Name: "panarchybot_tgbot_group_handler_total",
}, []string{"status"})

func TelegramBotPrivateMiddleware(next panarchybot.TelegramBotPrivateHandler) panarchybot.TelegramBotPrivateHandler {
	return func(ctx context.Context, state db.State, upd *models.Update, l *slog.Logger) error {
		err := next(ctx, state, upd, l)
		counterTGBotPrivateHandler.WithLabelValues(getStatus(err)).Inc()
		return err
	}
}

func TelegramBotGroupMiddleware(next panarchybot.TelegramBotGroupHandler) panarchybot.TelegramBotGroupHandler {
	return func(ctx context.Context, upd *models.Update, l *slog.Logger) error {
		err := next(ctx, upd, l)
		counterTGBotGroupHandler.WithLabelValues(getStatus(err)).Inc()
		return err
	}
}
