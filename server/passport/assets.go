package passport

import (
	"context"
	"server"

	"github.com/ninja-software/terror/v2"
)

type SuccessResponse struct {
	IsSuccess bool `json:"payload"`
}

// AssetFreeze tell passport to freeze user's assets
func (pp *Passport) AssetFreeze(ctx context.Context, txID string, assetTokenID uint64) error {
	replyChannel := make(chan []byte)
	errChan := make(chan error)

	pp.send <- &Request{
		ReplyChannel: replyChannel,
		ErrChan:      errChan,
		Message: &Message{
			Key: "SUPREMACY:ASSET:FREEZE",
			Payload: struct {
				AssetTokenID uint64 `json:"assetTokenID"`
			}{
				AssetTokenID: assetTokenID,
			},
			TransactionID: txID,
			context:       ctx,
		},
	}

	for {
		select {
		case <-replyChannel:
			return nil
		case err := <-errChan:
			return terror.Error(err)
		}
	}
}

// AssetLock tell passport to lock user's assets
func (pp *Passport) AssetLock(ctx context.Context, txID string, assetTokenIDs []uint64) error {
	replyChannel := make(chan []byte)
	errChan := make(chan error)

	pp.send <- &Request{
		ReplyChannel: replyChannel,
		ErrChan:      errChan,
		Message: &Message{
			Key: "SUPREMACY:ASSET:LOCK",
			Payload: struct {
				AssetTokenIDs []uint64 `json:"assetTokenIDs"`
			}{
				AssetTokenIDs: assetTokenIDs,
			},
			TransactionID: txID,
			context:       ctx,
		},
	}

	for {
		select {
		case <-replyChannel:
			return nil
		case err := <-errChan:
			return terror.Error(err)
		}
	}
}

// AssetRelease tell passport to release user's asset
func (pp *Passport) AssetRelease(ctx context.Context, txID string, releasedAssets []*server.WarMachineNFT) {
	pp.send <- &Request{
		Message: &Message{
			Key: "SUPREMACY:ASSET:RELEASE",
			Payload: struct {
				ReleasedAssets []*server.WarMachineNFT `json:"releasedAssets"`
			}{
				ReleasedAssets: releasedAssets,
			},
			context: ctx,
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
func (pp *Passport) WarMachineQueuePositionBroadcast(ctx context.Context, uwm []*UserWarMachineQueuePosition) {
	pp.send <- &Request{
		Message: &Message{
			Key: "SUPREMACY:WAR:MACHINE:QUEUE:POSITION",
			Payload: struct {
				UserWarMachineQueuePosition []*UserWarMachineQueuePosition `json:"userWarMachineQueuePosition"`
			}{
				UserWarMachineQueuePosition: uwm,
			},
			context: ctx,
		},
	}
}

// WarMachineQueue
func (pp *Passport) WarMachineQueuePositionClear(ctx context.Context, txID string, factionID server.FactionID) {
	pp.send <- &Request{
		Message: &Message{
			Key: "SUPREMACY:WAR:MACHINE:QUEUE:POSITION:CLEAR",
			Payload: struct {
				FactionID server.FactionID `json:"factionID"`
			}{
				FactionID: factionID,
			},
			context: ctx,
		},
	}
}
