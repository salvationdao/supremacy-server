package passport

import (
	"context"
	"encoding/json"
	"server"

	"github.com/ninja-software/terror/v2"
)

type SuccessResponse struct {
	IsSuccess bool `json:"payload"`
}

// AssetFreeze tell passport to freeze user's assets
func (pp *Passport) AssetFreeze(ctx context.Context, txID string, assetTokenID uint64) error {
	ctx, cancel := context.WithCancel(ctx)

	replyChannel := make(chan []byte)

	pp.send <- &Request{
		ReplyChannel: replyChannel,
		Message: &Message{
			Key: "SUPREMACY:ASSET:FREEZE",
			Payload: struct {
				AssetTokenID uint64 `json:"assetTokenID"`
			}{
				AssetTokenID: assetTokenID,
			},
			TransactionID: txID,
			context:       ctx,
			cancel:        cancel,
		},
	}

	msg := <-replyChannel
	resp := &SuccessResponse{}
	err := json.Unmarshal(msg, resp)
	if err != nil {
		return terror.Error(err)
	}

	if !resp.IsSuccess {
		return terror.Error(terror.ErrInvalidInput, "Unable to freeze passport asset")
	}

	return nil
}

// AssetLock tell passport to lock user's assets
func (pp *Passport) AssetLock(ctx context.Context, txID string, assetTokenIDs []uint64) error {
	ctx, cancel := context.WithCancel(ctx)

	replyChannel := make(chan []byte)

	pp.send <- &Request{
		ReplyChannel: replyChannel,
		Message: &Message{
			Key: "SUPREMACY:ASSET:LOCK",
			Payload: struct {
				AssetTokenIDs []uint64 `json:"assetTokenIDs"`
			}{
				AssetTokenIDs: assetTokenIDs,
			},
			TransactionID: txID,
			context:       ctx,
			cancel:        cancel,
		},
	}

	msg := <-replyChannel
	resp := &SuccessResponse{}
	err := json.Unmarshal(msg, resp)
	if err != nil {
		return terror.Error(err)
	}

	if !resp.IsSuccess {
		return terror.Error(terror.ErrInvalidInput, "Unable to lock passport asset")
	}

	return nil
}

// AssetRelease tell passport to release user's asset
func (pp *Passport) AssetRelease(ctx context.Context, txID string, releasedAssets []*server.WarMachineNFT) {
	ctx, cancel := context.WithCancel(ctx)

	pp.send <- &Request{
		Message: &Message{
			Key: "SUPREMACY:ASSET:RELEASE",
			Payload: struct {
				ReleasedAssets []*server.WarMachineNFT `json:"releasedAssets"`
			}{
				ReleasedAssets: releasedAssets,
			},
			TransactionID: txID,
			context:       ctx,
			cancel:        cancel,
		},
	}
}

type UserWarMachineQueuePosition struct {
	UserID                   server.UserID              `json:"userID"`
	WarMachineQueuePositions []*WarMachineQueuePosition `json:"warMachineQueuePositions"`
}

type WarMachineQueuePosition struct {
	WarMachineNFT *server.WarMachineNFT `json:"warMachineNFT"`
	Position      int                   `json:"position"`
}

// WarMachineQueue
func (pp *Passport) WarMachineQueuePosition(ctx context.Context, txID string, uwm []*UserWarMachineQueuePosition) {
	ctx, cancel := context.WithCancel(ctx)

	pp.send <- &Request{
		Message: &Message{
			Key: "SUPREMACY:WAR:MACHINE:QUEUE:POSITION",
			Payload: struct {
				UserWarMachineQueuePosition []*UserWarMachineQueuePosition `json:"userWarMachineQueuePosition"`
			}{
				UserWarMachineQueuePosition: uwm,
			},
			TransactionID: txID,
			context:       ctx,
			cancel:        cancel,
		},
	}
}

// WarMachineQueue
func (pp *Passport) WarMachineQueuePositionClear(ctx context.Context, txID string, factionID server.FactionID) {
	ctx, cancel := context.WithCancel(ctx)

	pp.send <- &Request{
		Message: &Message{
			Key: "SUPREMACY:WAR:MACHINE:QUEUE:POSITION:CLEAR",
			Payload: struct {
				FactionID server.FactionID `json:"factionID"`
			}{
				FactionID: factionID,
			},
			TransactionID: txID,
			context:       ctx,
			cancel:        cancel,
		},
	}
}
