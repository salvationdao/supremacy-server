package xsyn_rpcclient

import (
	"server/gamelog"
	"strings"
)

// SpendSupMessage tells the passport to make a transfer
func (pp *XsynXrpcClient) SpendSupMessage(req SpendSupsReq) (string, error) {
	req.ApiKey = pp.ApiKey
	resp := &SpendSupsResp{}
	err := pp.XrpcClient.Call("S.SupremacySpendSupsHandler", req, resp)
	if err != nil {
		// TODO: create a error type to check
		if !strings.Contains(err.Error(), "not enough funds") {
			gamelog.L.Err(err).Str("method", "SupremacySpendSupsHandler").Msg("rpc error")
		}
		return "", err
	}

	return resp.TransactionID, nil
}

// RefundSupsMessage tells the passport to refund a transaction
func (pp *XsynXrpcClient) RefundSupsMessage(transactionID string) (string, error) {
	resp := &RefundTransactionResp{}
	err := pp.XrpcClient.Call("S.RefundTransaction", &RefundTransactionReq{
		ApiKey:        pp.ApiKey,
		TransactionID: transactionID,
	}, resp)
	if err != nil {
		gamelog.L.Err(err).Str("method", "RefundTransaction").Msg("rpc error")
		return "", err
	}

	return resp.TransactionID, nil
}
