package server

import "github.com/gofrs/uuid"

type Faction struct {
	ID       FactionID `json:"id"`
	Label    string    `json:"label"`
	ImageUrl string    `json:"imageUrl"`
	Colour   string    `json:"colour"`
}

type FactionActionType string

const (
	FactionActionTypeAirStrike FactionActionType = "AIRSTRIKE"
	FactionActionTypeNuke      FactionActionType = "NUKE"
	FactionActionTypeHealing   FactionActionType = "HEALING"
)

type FactionAction struct {
	ID                FactionActionID   `json:"id"`
	Label             string            `json:"label"`
	Type              FactionActionType `json:"type"`
	Colour            string            `json:"colour"`
	SupremacyCoinCost int               `json:"supremacyCoinCost"`
	ImageUrl          string            `json:"imageUrl"`
}

/**************
* Fake Action *
**************/

var FactionActions = []*FactionAction{
	{
		ID:                FactionActionID(uuid.Must(uuid.NewV4())),
		Label:             "AIRSTRIKE",
		Type:              FactionActionTypeAirStrike,
		Colour:            "#428EC1",
		SupremacyCoinCost: 60,
		ImageUrl:          "https://i.pinimg.com/originals/b1/92/4d/b1924dce177345b5485bb5490ab3441f.jpg",
	},
	{
		ID:                FactionActionID(uuid.Must(uuid.NewV4())),
		Label:             "NUKE",
		Type:              FactionActionTypeNuke,
		Colour:            "#C24242",
		SupremacyCoinCost: 60,
		ImageUrl:          "https://images2.minutemediacdn.com/image/upload/c_crop,h_1126,w_2000,x_0,y_83/f_auto,q_auto,w_1100/v1555949079/shape/mentalfloss/581049-mesut_zengin-istock-1138195821.jpg",
	},
	{
		ID:                FactionActionID(uuid.Must(uuid.NewV4())),
		Label:             "HEAL",
		Type:              FactionActionTypeHealing,
		Colour:            "#30B07D",
		SupremacyCoinCost: 60,
		ImageUrl:          "https://i.pinimg.com/originals/ed/2f/9b/ed2f9b6e66b9efefa84d1ee423c718f0.png",
	},
}
