package syndicate

import (
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/ninja-software/terror/v2"
	"github.com/shopspring/decimal"
	"github.com/volatiletech/null/v8"
	"server"
	"server/db"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"server/xsyn_rpcclient"
	"sync"
	"time"
)

type AccountSystem struct {
	syndicate *Syndicate
	isLocked  bool
	sync.Mutex
}

func newAccountSystem(s *Syndicate) *AccountSystem {
	f := &AccountSystem{
		syndicate: s,
		isLocked:  false,
	}

	return f
}

func (as *AccountSystem) liquidate() error {
	as.Lock()
	defer as.Unlock()

	as.isLocked = true

	syndicateUUID := uuid.FromStringOrNil(as.syndicate.ID)

	fund := as.syndicate.system.Passport.UserBalanceGet(syndicateUUID)

	if fund.Equal(decimal.Zero) {
		gamelog.L.Debug().Msg("No fund in syndicate to distribute")
		return nil
	}

	// taxed
	taxRatio := db.GetDecimalWithDefault(db.KeyDecentralisedAutonomousSyndicateTax, decimal.NewFromFloat(0.025))
	if as.syndicate.Type == boiler.SyndicateTypeCORPORATION {
		taxRatio = db.GetDecimalWithDefault(db.KeyCorporationSyndicateTax, decimal.NewFromFloat(0.1))
	}

	tax := fund.Mul(taxRatio)

	_, err := as.syndicate.system.Passport.SpendSupMessage(xsyn_rpcclient.SpendSupsReq{
		FromUserID:           syndicateUUID,
		ToUserID:             uuid.Must(uuid.FromString(server.SupremacyGameUserID)),
		Amount:               tax.String(),
		TransactionReference: server.TransactionReference(fmt.Sprintf("liquidate_syndicate_tax:%s|%s|%d", as.syndicate.Type, as.syndicate.ID, time.Now().UnixNano())),
		Group:                string(server.TransactionGroupSupremacy),
		SubGroup:             string(server.TransactionGroupSyndicate),
		Description:          fmt.Sprintf("Tax for liquidating %s syndicate (%s%%): %s", as.syndicate.Type, taxRatio.String(), as.syndicate.ID),
		NotSafe:              true,
	})
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to tax syndicate for liquidation.")
		return terror.Error(err, "Failed to tax liquidation")
	}

	remainBalance := fund.Sub(tax)

	// equally distribute fund to all the remaining members
	members, err := as.syndicate.Players(
		boiler.PlayerWhere.SyndicateID.EQ(null.StringFrom(as.syndicate.ID)),
	).All(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Interface("syndicate id", as.syndicate.ID).Err(err).Msg("Failed to load syndicate members.")
		return terror.Error(err, "Failed to load syndicate members.")
	}

	fundPerMember := remainBalance.Div(decimal.NewFromInt(int64(len(members))))

	for _, m := range members {
		// distribute fun to remaining members
		transaction := xsyn_rpcclient.SpendSupsReq{
			FromUserID:           syndicateUUID,
			ToUserID:             uuid.Must(uuid.FromString(m.ID)),
			Amount:               fundPerMember.String(),
			TransactionReference: server.TransactionReference(fmt.Sprintf("remain_fund_after_liquidate_syndicate:%s|%s|%d", as.syndicate.Type, as.syndicate.ID, time.Now().UnixNano())),
			Group:                string(server.TransactionGroupSupremacy),
			SubGroup:             string(server.TransactionGroupSyndicate),
			Description:          fmt.Sprintf("Liquidated syndicate fund for remaining members: %s", as.syndicate.ID),
			NotSafe:              true,
		}
		_, err := as.syndicate.system.Passport.SpendSupMessage(transaction)
		if err != nil {
			gamelog.L.Error().Interface("transaction", transaction).Err(err).Msg("Failed to distribute remain fund to remaining members.")
			return terror.Error(err, "Failed to distribute remain fund to remaining members.")
		}
	}

	return nil
}

func (as *AccountSystem) receiveFund(fromID string, fund decimal.Decimal, reference server.TransactionReference, description string) error {
	as.Lock()
	defer as.Unlock()

	if as.isLocked {
		return terror.Error(fmt.Errorf("syndicate account is locked"), "Syndicate account is locked")
	}

	// distribute fun to remaining members
	transaction := xsyn_rpcclient.SpendSupsReq{
		FromUserID:           uuid.FromStringOrNil(fromID),
		ToUserID:             uuid.FromStringOrNil(as.syndicate.ID),
		Amount:               fund.String(),
		TransactionReference: reference,
		Group:                string(server.TransactionGroupSupremacy),
		SubGroup:             string(server.TransactionGroupSyndicate),
		Description:          description,
		NotSafe:              true,
	}
	_, err := as.syndicate.system.Passport.SpendSupMessage(transaction)
	if err != nil {
		gamelog.L.Error().Interface("transaction", transaction).Err(err).Msg("Failed to receive fund.")
		return terror.Error(err, "Failed to receive fund.")
	}

	return nil
}

func (as *AccountSystem) transferFund(toID string, fund decimal.Decimal, reference server.TransactionReference, description string) error {
	as.Lock()
	defer as.Unlock()

	if as.isLocked {
		return terror.Error(fmt.Errorf("syndicate account is locked"), "Syndicate account is locked")
	}

	// distribute fun to remaining members
	transaction := xsyn_rpcclient.SpendSupsReq{
		FromUserID:           uuid.FromStringOrNil(as.syndicate.ID),
		ToUserID:             uuid.FromStringOrNil(toID),
		Amount:               fund.String(),
		TransactionReference: reference,
		Group:                string(server.TransactionGroupSupremacy),
		SubGroup:             string(server.TransactionGroupSyndicate),
		Description:          description,
		NotSafe:              true,
	}
	_, err := as.syndicate.system.Passport.SpendSupMessage(transaction)
	if err != nil {
		gamelog.L.Error().Interface("transaction", transaction).Err(err).Msg("Failed to transfer fund.")
		return terror.Error(err, "Failed to transfer fund.")
	}

	return nil
}
