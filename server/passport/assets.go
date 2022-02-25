package passport

import (
	"context"
	"server"

	"github.com/gofrs/uuid"
)

type SuccessResponse struct {
	IsSuccess bool `json:"payload"`
}

// AssetFreeze tell passport to freeze user's assets
func (pp *Passport) AssetFreeze(ctx context.Context, assetTokenID uint64) error {
	pp.send <- &Message{
		Key: "SUPREMACY:ASSET:FREEZE",
		Payload: struct {
			AssetTokenID uint64 `json:"assetTokenID"`
		}{
			AssetTokenID: assetTokenID,
		},
		TransactionID: uuid.Must(uuid.NewV4()).String(),
	}
	return nil
}

// AssetLock tell passport to lock user's assets
func (pp *Passport) AssetLock(ctx context.Context, assetTokenIDs []uint64) error {
	pp.send <- &Message{
		Key: "SUPREMACY:ASSET:LOCK",
		Payload: struct {
			AssetTokenIDs []uint64 `json:"assetTokenIDs"`
		}{
			AssetTokenIDs: assetTokenIDs,
		},
		TransactionID: uuid.Must(uuid.NewV4()).String(),
	}
	return nil
}

// AssetRelease tell passport to release user's asset
func (pp *Passport) AssetRelease(ctx context.Context, releasedAssets []*server.WarMachineMetadata) {
	pp.send <- &Message{
		Key: "SUPREMACY:ASSET:RELEASE",
		Payload: struct {
			ReleasedAssets []*server.WarMachineMetadata `json:"releasedAssets"`
		}{
			ReleasedAssets: releasedAssets,
		},
	}
}

type UserWarMachineQueuePosition struct {
	UserID                   server.UserID              `json:"userID"`
	WarMachineQueuePositions []*WarMachineQueuePosition `json:"warMachineQueuePositions"`
}

type WarMachineQueuePosition struct {
	WarMachineMetadata *server.WarMachineMetadata `json:"warMachineMetadata"`
	Position           int                        `json:"position"`
}

func (pp *Passport) WarMachineQueuePositionBroadcast(uwm []*UserWarMachineQueuePosition) {
	pp.send <- &Message{
		Key: "SUPREMACY:WAR:MACHINE:QUEUE:POSITION",
		Payload: struct {
			UserWarMachineQueuePosition []*UserWarMachineQueuePosition `json:"userWarMachineQueuePosition"`
		}{
			UserWarMachineQueuePosition: uwm,
		},
	}
}

func (pp *Passport) AbilityUpdateTargetPrice(abilityTokenID, warMachineTokenID uint64, supsCost string) {
	pp.send <- &Message{
		Key: "SUPREMACY:ABILITY:TARGET:PRICE:UPDATE",
		Payload: struct {
			AbilityTokenID    uint64 `json:"abilityTokenID"`
			WarMachineTokenID uint64 `json:"warMachineTokenID"`
			SupsCost          string `json:"supsCost"`
		}{
			AbilityTokenID:    abilityTokenID,
			WarMachineTokenID: warMachineTokenID,
			SupsCost:          supsCost,
		},
	}
}

// AssetInsurancePay tell passport to pay insurance for battle asset
func (pp *Passport) AssetInsurancePay(userID server.UserID, factionID server.FactionID, amount server.BigInt, txRef server.TransactionReference) error {
	pp.send <- &Message{
		Key: "SUPREMACY:PAY_ASSET_INSURANCE",
		Payload: struct {
			UserID               server.UserID               `json:"userID"`
			FactionID            server.FactionID            `json:"factionID"`
			Amount               server.BigInt               `json:"amount"`
			TransactionReference server.TransactionReference `json:"transactionReference"`
		}{
			UserID:               userID,
			FactionID:            factionID,
			Amount:               amount,
			TransactionReference: txRef,
		},
		TransactionID: uuid.Must(uuid.NewV4()).String(),
	}

	return nil
}

// AssetContractRewardRedeem redeem faction contract reward
func (pp *Passport) AssetContractRewardRedeem(ctx context.Context, userID server.UserID, factionID server.FactionID, amount server.BigInt, txRef server.TransactionReference) error {
	pp.send <- &Message{
		Key: "SUPREMACY:REDEEM_FACTION_CONTRACT_REWARD",
		Payload: struct {
			UserID               server.UserID               `json:"userID"`
			FactionID            server.FactionID            `json:"factionID"`
			Amount               server.BigInt               `json:"amount"`
			TransactionReference server.TransactionReference `json:"transactionReference"`
		}{
			UserID:               userID,
			FactionID:            factionID,
			Amount:               amount,
			TransactionReference: txRef,
		},
		TransactionID: uuid.Must(uuid.NewV4()).String(),
	}
	return nil
}
