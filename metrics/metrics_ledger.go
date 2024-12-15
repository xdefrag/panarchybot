package metrics

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stellar/go/keypair"
	"github.com/xdefrag/panarchybot"
)

var counterNewAccount = prometheus.NewCounterVec(prometheus.CounterOpts{
	Name: "panarchybot_new_account_total",
}, []string{"status"})

var counterGetBalance = prometheus.NewCounterVec(prometheus.CounterOpts{
	Name: "panarchybot_get_balance_total",
}, []string{"status"})

var counterSend = prometheus.NewCounterVec(prometheus.CounterOpts{
	Name: "panarchybot_send_total",
}, []string{"status"})

type LedgerWrapper struct {
	ledger panarchybot.Ledger
}

// CreateAccount implements panarchybot.Ledger.
func (l *LedgerWrapper) CreateAccount(ctx context.Context) (*keypair.Full, error) {
	res, err := l.ledger.CreateAccount(ctx)
	counterNewAccount.WithLabelValues(getStatus(err)).Inc()
	return res, err
}

// GetBalance implements panarchybot.Ledger.
func (l *LedgerWrapper) GetBalance(ctx context.Context, address string) (string, error) {
	res, err := l.ledger.GetBalance(ctx, address)
	counterGetBalance.WithLabelValues(getStatus(err)).Inc()
	return res, err
}

// Send implements panarchybot.Ledger.
func (l *LedgerWrapper) Send(ctx context.Context, fromSeed string, toAddress string, amount string, opts ...panarchybot.SendOption) (string, error) {
	res, err := l.ledger.Send(ctx, fromSeed, toAddress, amount, opts...)
	counterSend.WithLabelValues(getStatus(err)).Inc()
	return res, err
}

func NewLedgerWrapper(ledger panarchybot.Ledger) *LedgerWrapper {
	prometheus.MustRegister(
		counterNewAccount,
		counterGetBalance,
		counterSend,
	)

	return &LedgerWrapper{
		ledger: ledger,
	}
}

var _ panarchybot.Ledger = (*LedgerWrapper)(nil)
