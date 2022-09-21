package xsyn_rpcclient

import (
	"server/gamelog"

	"github.com/shopspring/decimal"
)

// GetCurrentSupPrice retrieves the current SUP price in USD from passport.
func (pp *XsynXrpcClient) GetCurrentSupPrice() (decimal.Decimal, error) {
	resp := &GetCurrentSupPriceResp{}
	err := pp.XrpcClient.Call("S.GetCurrentSupPrice", &GetCurrentSupPriceReq{}, resp)
	if err != nil {
		gamelog.L.Err(err).Str("method", "RefundTransaction").Msg("rpc error")
		return decimal.Zero, err
	}

	return resp.PriceUSD, nil
}
