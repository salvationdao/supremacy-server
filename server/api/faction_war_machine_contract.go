package api

//////////////////
// Asset Repair //
//////////////////

// type RepairQueue map[string]bool

// func (api *API) InitialAssetRepairCenter() {
// 	api.fastAssetRepairCenter = make(chan func(RepairQueue))
// 	fastRepairQueue := make(RepairQueue)

// 	// get all the unrepair
// 	hashes, err := db.AssetUnrepairList(context.Background(), api.Conn)
// 	if err != nil {
// 		api.Log.Err(err).Msg("failed to get initail unrepair war machines from db when initialise")
// 	}

// 	// chuck war machines in fast route repair
// 	for _, hash := range hashes {
// 		fastRepairQueue[hash] = true
// 	}

// 	go func() {
// 		for fn := range api.fastAssetRepairCenter {
// 			fn(fastRepairQueue)
// 		}
// 	}()
// 	api.startRepairTicker(RepairTypeFast)

// 	api.standardAssetRepairCenter = make(chan func(RepairQueue))
// 	standerRepairQueue := make(RepairQueue)
// 	go func() {
// 		for fn := range api.standardAssetRepairCenter {
// 			fn(standerRepairQueue)
// 		}
// 	}()
// 	api.startRepairTicker(RepairTypeStandard)
// }

// type RepairType string

// const (
// 	RepairTypeFast     RepairType = "FAST"
// 	RepairTypeStandard RepairType = "STANDARD"
// )

// func (api *API) RegisterRepairCenter(rt RepairType, assetHash string) {
// 	switch rt {
// 	case RepairTypeFast:
// 		api.fastAssetRepairCenter <- func(rq RepairQueue) {
// 			rq[assetHash] = true
// 		}

// 	case RepairTypeStandard:
// 		api.standardAssetRepairCenter <- func(rq RepairQueue) {
// 			rq[assetHash] = true
// 		}
// 	}
// }

// func (api *API) startRepairTicker(rt RepairType) {
// 	tickSecond := 0
// 	TraceTitle := ""
// 	var repairCenter chan func(RepairQueue)
// 	switch rt {
// 	case RepairTypeFast:
// 		tickSecond = 18 // repair from 0 to 100 take 30 minutes
// 		TraceTitle = "Fast Repair Center"
// 		repairCenter = api.fastAssetRepairCenter
// 	case RepairTypeStandard:
// 		tickSecond = 864 // repair from 0 to 100 take 24 hours
// 		TraceTitle = "Standard Repair Center"
// 		repairCenter = api.standardAssetRepairCenter
// 	}

// 	// build tickle
// 	assetRepairCenter := tickle.New(TraceTitle, float64(tickSecond), func() (int, error) {
// 		errChan := make(chan error)
// 		repairCenter <- func(rq RepairQueue) {
// 			if len(rq) == 0 {
// 				errChan <- nil
// 				return
// 			}

// 			assetHashes := []string{}
// 			for hash := range rq {
// 				assetHashes = append(assetHashes, hash)
// 			}

// 			nfts, err := db.XsynAssetDurabilityBulkIncrement(context.Background(), api.Conn, assetHashes)
// 			if err != nil {
// 				errChan <- err
// 				return
// 			}

// 			// remove war machine which is completely repaired
// 			for _, nft := range nfts {
// 				if nft.Durability == 100 {
// 					delete(rq, nft.Hash)
// 				}
// 			}
// 			errChan <- nil
// 		}

// 		err := <-errChan
// 		if err != nil {
// 			return http.StatusInternalServerError, terror.Error(err)
// 		}
// 		return http.StatusOK, nil
// 	})
// 	assetRepairCenter.Log = log_helpers.NamedLogger(api.Log, TraceTitle)
// 	assetRepairCenter.DisableLogging = true

// 	assetRepairCenter.Start()
// }
