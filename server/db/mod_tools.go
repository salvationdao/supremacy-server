package db

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/volatiletech/null/v8"
	"server"
	"server/db/boiler"
	"server/gamedb"
	"time"

	"github.com/ninja-software/terror/v2"
	"github.com/shopspring/decimal"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type AdminToolResponse struct {
	User              *server.Player        `json:"user"`
	UserAssets        *AdminToolUserAsset   `json:"user_assets,omitempty"`
	BanHistory        []*AdminBanHistory    `json:"ban_history,omitempty"`
	ActiveBan         []*AdminBanHistory    `json:"active_ban, omitempty"`
	RecentChatHistory []*boiler.ChatHistory `json:"recent_chat_history,omitempty"`
	RelatedAccounts   []*server.Player      `json:"related_accounts,omitempty"`
}

type AdminBanHistory struct {
	ID                     string        `json:"id"`
	CreatedAt              time.Time     `json:"created_at"`
	Reason                 string        `json:"reason"`
	EndAt                  time.Time     `json:"end_at"`
	BannedAt               time.Time     `json:"banned_at"`
	BannedBy               server.Player `json:"banned_by"`
	ManuallyUnbanned       bool          `json:"manually_unbanned"`
	ManuallyUnbannedReason null.String   `json:"manually_unbanned_reason"`
}

type AdminToolUserAsset struct {
	Mechs []*server.Mech  `json:"mechs"`
	Sups  decimal.Decimal `json:"sups"`
}

func ModToolGetUserData(userID string, isAdmin bool, supsAmount decimal.Decimal) (*AdminToolResponse, error) {
	player, err := boiler.FindPlayer(gamedb.StdConn, userID)
	if err != nil {
		return nil, terror.Error(err, "Failed to find player for admin tool")
	}

	now := time.Now()

	playerBansHistory, err := boiler.PlayerBans(
		boiler.PlayerBanWhere.BannedPlayerID.EQ(userID),
		qm.Where(fmt.Sprintf(
			`(%s < ? OR %s NOTNULL)`,
			boiler.PlayerBanTableColumns.EndAt,
			boiler.PlayerBanTableColumns.ManuallyUnbanByID,
		), now),
		qm.OrderBy(fmt.Sprintf("%s DESC", boiler.PlayerBanTableColumns.CreatedAt)),
		qm.Load(boiler.PlayerBanRels.BannedBy),
	).All(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, terror.Error(err, "Failed to get player bans")
	}

	adminBanHistories := []*AdminBanHistory{}

	if len(playerBansHistory) > 0 {
		for _, pb := range playerBansHistory {
			adminBanHistory := &AdminBanHistory{
				ID:                     pb.ID,
				CreatedAt:              pb.CreatedAt,
				BannedAt:               pb.BannedAt,
				Reason:                 pb.Reason,
				EndAt:                  pb.EndAt,
				ManuallyUnbanned:       pb.ManuallyUnbanByID.Valid,
				ManuallyUnbannedReason: pb.ManuallyUnbanReason,
			}

			if pb.R != nil && pb.R.BannedBy != nil {
				adminBanHistory.BannedBy = *server.PlayerFromBoiler(pb.R.BannedBy)
			}

			adminBanHistories = append(adminBanHistories, adminBanHistory)
		}
	}

	playerActiveBans, err := boiler.PlayerBans(
		boiler.PlayerBanWhere.BannedPlayerID.EQ(userID),
		boiler.PlayerBanWhere.EndAt.GT(now),
		boiler.PlayerBanWhere.ManuallyUnbanAt.IsNull(),
		qm.OrderBy(fmt.Sprintf("%s DESC", boiler.PlayerBanTableColumns.CreatedAt)),
		qm.Load(boiler.PlayerBanRels.BannedBy),
	).All(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, terror.Error(err, "Failed to get player bans")
	}

	activeBans := []*AdminBanHistory{}

	if len(playerActiveBans) > 0 {
		for _, pb := range playerActiveBans {
			adminBanHistory := &AdminBanHistory{
				ID:                     pb.ID,
				CreatedAt:              pb.CreatedAt,
				BannedAt:               pb.BannedAt,
				Reason:                 pb.Reason,
				EndAt:                  pb.EndAt,
				ManuallyUnbanned:       pb.ManuallyUnbanByID.Valid,
				ManuallyUnbannedReason: pb.ManuallyUnbanReason,
			}

			if pb.R != nil && pb.R.BannedBy != nil {
				adminBanHistory.BannedBy = *server.PlayerFromBoiler(pb.R.BannedBy)
			}

			activeBans = append(activeBans, adminBanHistory)
		}
	}

	recentChatHistory, err := boiler.ChatHistories(
		boiler.ChatHistoryWhere.PlayerID.EQ(userID),
		qm.OrderBy(fmt.Sprintf("%s DESC", boiler.ChatHistoryTableColumns.CreatedAt)),
		qm.Limit(15),
	).All(gamedb.StdConn)
	if err != nil {
		return nil, terror.Error(err, "Failed to get recent chat history")
	}

	adminToolResponse := &AdminToolResponse{
		User:              server.PlayerFromBoiler(player),
		RecentChatHistory: recentChatHistory,
	}

	if len(adminBanHistories) > 0 {
		adminToolResponse.BanHistory = adminBanHistories
	}

	if len(playerActiveBans) > 0 {
		adminToolResponse.ActiveBan = activeBans
	}

	relatedAccouts, err := getPlayerRelatedAccounts(player.ID)
	if err != nil {
		return nil, terror.Error(err, "Failed to get related accounts")
	}

	adminToolResponse.RelatedAccounts = relatedAccouts

	if isAdmin {
		userAssets := &AdminToolUserAsset{
			Sups: supsAmount,
		}

		mechs, err := boiler.CollectionItems(
			boiler.CollectionItemWhere.ItemType.EQ(boiler.ItemTypeMech),
			boiler.CollectionItemWhere.OwnerID.EQ(player.ID),
		).All(gamedb.StdConn)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return nil, terror.Error(err, "Failed to find users mechs")
		}

		if mechs != nil {
			mechIDs := []string{}

			for _, mech := range mechs {
				mechIDs = append(mechIDs, mech.ItemID)
			}

			mechBriefs, err := Mechs(mechIDs...)
			if err != nil {
				return nil, terror.Error(err, "Failed to get owned mechs brief")
			}

			userAssets.Mechs = mechBriefs
		}

		adminToolResponse.UserAssets = userAssets
	}

	return adminToolResponse, nil
}

func getPlayerRelatedAccounts(userID string) ([]*server.Player, error) {
	rows, err := boiler.NewQuery(
		[]qm.QueryMod{
			qm.Select(fmt.Sprintf("TO_JSON(%s)", boiler.TableNames.Players)),
			qm.From(fmt.Sprintf(
				`(
						SELECT DISTINCT (%s) AS id
						FROM (SELECT %s FROM %s WHERE %s = '%s') _pf
						INNER JOIN %s ON %s = _pf.%s
					) p`,
				boiler.PlayerFingerprintTableColumns.PlayerID,
				boiler.PlayerFingerprintColumns.FingerprintID,
				boiler.TableNames.PlayerFingerprints,
				boiler.PlayerFingerprintColumns.PlayerID,
				userID,
				boiler.TableNames.PlayerFingerprints,
				boiler.PlayerFingerprintTableColumns.FingerprintID,
				boiler.PlayerFingerprintColumns.FingerprintID,
			)),
			qm.InnerJoin(fmt.Sprintf(
				"%s ON %s = p.%s",
				boiler.TableNames.Players,
				boiler.PlayerTableColumns.ID,
				boiler.PlayerColumns.ID,
			)),
		}...,
	).Query(gamedb.StdConn)
	if err != nil {
		return nil, err
	}

	relatedAccounts := []*server.Player{}

	for rows.Next() {
		relatedPlayer := &server.Player{}
		err = rows.Scan(&relatedPlayer)
		if err != nil {
			return nil, err
		}

		if relatedPlayer.ID == userID {
			continue
		}
		relatedAccounts = append(relatedAccounts, relatedPlayer)
	}

	return relatedAccounts, nil
}