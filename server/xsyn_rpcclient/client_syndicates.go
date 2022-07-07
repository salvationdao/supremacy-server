package xsyn_rpcclient

import "server/gamelog"

// SyndicateCreateHandler tells the passport to create a copy of the syndicate
func (pp *XsynXrpcClient) SyndicateCreateHandler(id string, foundedByID string, name string) error {
	err := pp.XrpcClient.Call("S.SyndicateCreateHandler", SyndicateCreateReq{
		ApiKey:      pp.ApiKey,
		ID:          id,
		FoundedByID: foundedByID,
		Name:        name,
	}, &SyndicateCreateResp{})
	if err != nil {
		gamelog.L.Err(err).Str("method", "SyndicateCreateHandler").Msg("rpc error")
		return err
	}

	return nil
}
