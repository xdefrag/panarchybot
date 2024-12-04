package stellar

import (
	"context"

	"github.com/stellar/go/clients/horizonclient"
)

func (s *Stellar) GetBalance(ctx context.Context, address string) (string, error) {
	page, err := s.cl.Accounts(horizonclient.AccountsRequest{
		Signer: address,
	})
	if err != nil {
		return "", err
	}

	if len(page.Embedded.Records) == 0 {
		return "", nil
	}

	return page.Embedded.Records[0].GetCreditBalance(s.cfg.Stellar.FundAccount.AssetCode, s.cfg.Stellar.FundAccount.AssetIssuer), nil
}
