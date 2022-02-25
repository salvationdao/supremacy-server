package passport

import (
	"context"
	"fmt"
	"server"

	"github.com/gofrs/uuid"
)

type SuccessResponse struct {
	IsSuccess bool `json:"payload"`
}

// AssetFreeze tell passport to freeze user's assets
func (pp *Passport) AssetFreeze(ctx context.Context, assetHash string) error {
	pp.send <- &Message{
		Key: "SUPREMACY:ASSET:FREEZE",
		Payload: struct {
			AssetHash string `json:"assetHash"`
		}{
			AssetHash: assetHash,
		},
		TransactionID: uuid.Must(uuid.NewV4()).String(),
	}
	return nil
}

// AssetLock tell passport to lock user's assets
func (pp *Passport) AssetLock(ctx context.Context, assetHashes []string) error {
	fmt.Println("888888888888888888888888888888888888888888888888888888888888888888")

	pp.send <- &Message{
		Key: "SUPREMACY:ASSET:LOCK",
		Payload: struct {
			AssetHashes []string `json:"assetHashes"`
		}{
			AssetHashes: assetHashes,
		},
		TransactionID: uuid.Must(uuid.NewV4()).String(),
	}

	fmt.Println("Enter 12333333333333333333333333333333333333333333333333333333")

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

func (pp *Passport) AbilityUpdateTargetPrice(abilityHash, warMachineHash string, supsCost string) {
	pp.send <- &Message{
		Key: "SUPREMACY:ABILITY:TARGET:PRICE:UPDATE",
		Payload: struct {
			AbilityHash    string `json:"abilityHash"`
			WarMachineHash string `json:"warMachineHash"`
			SupsCost       string `json:"supsCost"`
		}{
			AbilityHash:    abilityHash,
			WarMachineHash: warMachineHash,
			SupsCost:       supsCost,
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
