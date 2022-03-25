package rpcclient

import (
	"server/gamelog"

	"github.com/ninja-software/terror/v2"
)

// SpendSupMessage tells the passport to hold sups
func (pp *PassportXrpcClient) SpendSupMessage(req SpendSupsReq) (string, error) {
	req.ApiKey = pp.ApiKey
	resp := &SpendSupsResp{}
	err := pp.XrpcClient.Call("S.SupremacySpendSupsHandler", req, resp)
	if err != nil {
		gamelog.L.Err(err).Str("method", "SupremacySpendSupsHandler").Msg("rpc error")
		return "", terror.Error(err)
	}

	return resp.TXID, nil
}
