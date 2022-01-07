package battle_arena

import (
	"context"
	"fmt"
	"server"
	"server/db"
	"time"

	"github.com/gofrs/uuid"
	"github.com/ninja-software/terror/v2"
)

type ActionTriggerRequest struct {
	BattleID        server.BattleID
	FactionID       server.FactionID
	FactionActionID server.FactionActionID
	IsSuccess       bool
}

func (ba *BattleArena) FactionActionTrigger(tat *ActionTriggerRequest) error {
	go ba.fakeAnimation(tat.FactionID)

	// get action
	action := &server.FactionAction{}
	for _, actn := range server.FactionActions {
		if actn.ID == tat.FactionActionID {
			action = actn
		}
	}
	if action == nil {
		return terror.Error(fmt.Errorf("unable to find action %s", tat.FactionActionID))
	}

	ctx := context.Background()
	// save to database
	tx, err := ba.Conn.Begin(ctx)
	if err != nil {
		return terror.Error(err)
	}

	err = db.FactionActionTriggered(ctx, tx, tat.BattleID, server.FactionAbility{
		FactionID:  tat.FactionID,
		Action:     *action,
		Successful: tat.IsSuccess,
	})
	if err != nil {
		return terror.Error(err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return terror.Error(err)
	}

	return nil
}

func (ba *BattleArena) fakeAnimation(factionID server.FactionID) {
	// want second
	i := 5
	for i > 0 {
		fmt.Println("wait", i, "seconds for animation to end")
		time.Sleep(1 * time.Second)
		i--
	}
	fmt.Println("wait", i, "seconds for animation to end")
	fmt.Println("----------------------------------")
	fmt.Println("Restart the voting cycle")

	ba.Events.Trigger(context.Background(), Event(fmt.Sprintf("%s:%s", factionID, EventAnamationEnd)), nil)
}

func (ba *BattleArena) FakeWarMachinePositionUpdate() {
	i := 1
	for {
		for _, warMachine := range ba.battle.WarMachines {
			// do update
			scale := 1
			if i%2 == 0 {
				scale = -1
			}

			warMachine.Rotation = i % 360
			warMachine.Position.X += 10 * scale
			warMachine.Position.Y -= 10 * scale
		}

		// broadcast
		ba.Events.Trigger(context.Background(), EventWarMachinePositionChanged, &EventData{
			BattleArena: ba.battle,
		})

		time.Sleep(250 * time.Millisecond)
		i++
	}

}

/*************
* Dummy Data *
*************/

var factionActions = []*server.FactionAction{
	{
		ID:                server.FactionActionID(uuid.Must(uuid.NewV4())),
		Label:             "AIRSTRIKE",
		Type:              server.FactionActionTypeAirStrike,
		Colour:            "#428EC1",
		SupremacyCoinCost: 60,
		ImageUrl:          "https://i.pinimg.com/originals/b1/92/4d/b1924dce177345b5485bb5490ab3441f.jpg",
	},
	{
		ID:                server.FactionActionID(uuid.Must(uuid.NewV4())),
		Label:             "NUKE",
		Type:              server.FactionActionTypeNuke,
		Colour:            "#C24242",
		SupremacyCoinCost: 60,
		ImageUrl:          "https://images2.minutemediacdn.com/image/upload/c_crop,h_1126,w_2000,x_0,y_83/f_auto,q_auto,w_1100/v1555949079/shape/mentalfloss/581049-mesut_zengin-istock-1138195821.jpg",
	},
	{
		ID:                server.FactionActionID(uuid.Must(uuid.NewV4())),
		Label:             "HEAL",
		Type:              server.FactionActionTypeHealing,
		Colour:            "#30B07D",
		SupremacyCoinCost: 60,
		ImageUrl:          "https://i.pinimg.com/originals/ed/2f/9b/ed2f9b6e66b9efefa84d1ee423c718f0.png",
	},
}
