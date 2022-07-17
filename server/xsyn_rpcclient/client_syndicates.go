package xsyn_rpcclient

import "server/gamelog"

// SyndicateCreateHandler tells the passport to create a copy of the syndicate
func (pp *XsynXrpcClient) SyndicateCreateHandler(id string, name string, foundedByID string) error {
	err := pp.XrpcClient.Call("S.SyndicateCreateHandler", SyndicateCreateReq{
		ApiKey:      pp.ApiKey,
		SyndicateID: id,
		FoundedByID: foundedByID,
		Name:        name,
	}, &SyndicateCreateResp{})
	if err != nil {
		gamelog.L.Err(err).Str("method", "SyndicateCreateHandler").Msg("rpc error")
		return err
	}

	return nil
}

// SyndicateNameChangeHandler tells the passport to change syndicate name
func (pp *XsynXrpcClient) SyndicateNameChangeHandler(id string, name string) error {
	err := pp.XrpcClient.Call("S.SyndicateNameChangeHandler", SyndicateNameCreateReq{
		ApiKey:      pp.ApiKey,
		SyndicateID: id,
		Name:        name,
	}, &SyndicateCreateResp{})
	if err != nil {
		gamelog.L.Err(err).Str("method", "SyndicateNameChangeHandler").Msg("rpc error")
		return err
	}

	return nil
}

// SyndicateLiquidateHandler tells the passport to archive syndicate
func (pp *XsynXrpcClient) SyndicateLiquidateHandler(id string) error {
	err := pp.XrpcClient.Call("S.SyndicateLiquidateHandler", SyndicateLiquidateReq{
		ApiKey:      pp.ApiKey,
		SyndicateID: id,
	}, &SyndicateCreateResp{})
	if err != nil {
		gamelog.L.Err(err).Str("method", "SyndicateLiquidateHandler").Msg("rpc error")
		return err
	}

	return nil
}
