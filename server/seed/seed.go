package seed

import (
	"context"
	"fmt"
	"server"
	"server/db"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/ninja-software/terror/v2"
)

type Seeder struct {
	Conn *pgxpool.Pool
}

// NewSeeder returns a new Seeder
func NewSeeder(conn *pgxpool.Pool) *Seeder {
	s := &Seeder{conn}
	return s
}

// Run for database spin up
func (s *Seeder) Run() error {
	// seed review icons
	fmt.Println("Seed game maps")
	ctx := context.Background()
	err := gameMaps(ctx, s.Conn)
	if err != nil {
		return terror.Error(err)
	}

	fmt.Println("seed factions")
	_, err = s.factions(ctx)
	if err != nil {
		return terror.Error(err)
	}

	fmt.Println("Seed faction abilities")
	err = factionAbilities(ctx, s.Conn)
	if err != nil {
		return terror.Error(err)
	}

	fmt.Println("Seed complete!")

	return nil
}

func gameMaps(ctx context.Context, conn *pgxpool.Pool) error {
	for _, gameMap := range GameMaps {
		err := db.GameMapCreate(ctx, conn, gameMap)
		if err != nil {
			fmt.Println(err)
			return terror.Error(err)
		}
	}
	return nil
}

var FactionIDRedMountain = server.FactionID(uuid.Must(uuid.FromString("98bf7bb3-1a7c-4f21-8843-458d62884060")))
var FactionIDBoston = server.FactionID(uuid.Must(uuid.FromString("7c6dde21-b067-46cf-9e56-155c88a520e2")))
var FactionIDZaibatsu = server.FactionID(uuid.Must(uuid.FromString("880db344-e405-428d-84e5-6ebebab1fe6d")))

var SharedAbilityCollections = []*server.BattleAbility{
	{
		Label:                  "AIRSTRIKE",
		CooldownDurationSecond: 20,
	},
	{
		Label:                  "NUKE",
		CooldownDurationSecond: 30,
	},
	{
		Label:                  "SYNDICATE CHOICE",
		CooldownDurationSecond: 15,
	},
}

var SharedFactionAbilities = []*server.GameAbility{
	// FactionIDZaibatsu
	{
		Label:               "AIRSTRIKE",
		FactionID:           FactionIDZaibatsu,
		GameClientAbilityID: 0,
		Colour:              "#428EC1",
		ImageUrl:            "https://i.pinimg.com/originals/b1/92/4d/b1924dce177345b5485bb5490ab3441f.jpg",
		SupsCost:            "0",
	},
	{
		Label:               "NUKE",
		FactionID:           FactionIDZaibatsu,
		GameClientAbilityID: 1,
		Colour:              "#C24242",
		ImageUrl:            "https://images2.minutemediacdn.com/image/upload/c_crop,h_1126,w_2000,x_0,y_83/f_auto,q_auto,w_1100/v1555949079/shape/mentalfloss/581049-mesut_zengin-istock-1138195821.jpg",
		SupsCost:            "0",
	},
	{
		Label:               "SYNDICATE CHOICE",
		FactionID:           FactionIDZaibatsu,
		GameClientAbilityID: 2,
		Colour:              "#30B07D",
		ImageUrl:            "https://i.pinimg.com/originals/ed/2f/9b/ed2f9b6e66b9efefa84d1ee423c718f0.png",
		SupsCost:            "0",
	},
	// FactionIDBoston
	{
		Label:               "AIRSTRIKE",
		GameClientAbilityID: 3,
		FactionID:           FactionIDBoston,
		Colour:              "#428EC1",
		ImageUrl:            "https://i.pinimg.com/originals/b1/92/4d/b1924dce177345b5485bb5490ab3441f.jpg",
		SupsCost:            "0",
	},
	{
		Label:               "NUKE",
		FactionID:           FactionIDBoston,
		GameClientAbilityID: 4,
		Colour:              "#C24242",
		ImageUrl:            "https://images2.minutemediacdn.com/image/upload/c_crop,h_1126,w_2000,x_0,y_83/f_auto,q_auto,w_1100/v1555949079/shape/mentalfloss/581049-mesut_zengin-istock-1138195821.jpg",
		SupsCost:            "0",
	},
	{
		Label:               "SYNDICATE CHOICE",
		FactionID:           FactionIDBoston,
		GameClientAbilityID: 5,
		Colour:              "#30B07D",
		ImageUrl:            "https://i.pinimg.com/originals/ed/2f/9b/ed2f9b6e66b9efefa84d1ee423c718f0.png",
		SupsCost:            "0",
	},
	// FactionIDRedMountain
	{
		Label:               "AIRSTRIKE",
		FactionID:           FactionIDRedMountain,
		GameClientAbilityID: 6,
		Colour:              "#428EC1",
		ImageUrl:            "https://i.pinimg.com/originals/b1/92/4d/b1924dce177345b5485bb5490ab3441f.jpg",
		SupsCost:            "0",
	},
	{
		Label:               "NUKE",
		FactionID:           FactionIDRedMountain,
		GameClientAbilityID: 7,
		Colour:              "#C24242",
		ImageUrl:            "https://images2.minutemediacdn.com/image/upload/c_crop,h_1126,w_2000,x_0,y_83/f_auto,q_auto,w_1100/v1555949079/shape/mentalfloss/581049-mesut_zengin-istock-1138195821.jpg",
		SupsCost:            "0",
	},
	{
		Label:               "SYNDICATE CHOICE",
		FactionID:           FactionIDRedMountain,
		GameClientAbilityID: 8,
		Colour:              "#30B07D",
		ImageUrl:            "https://i.pinimg.com/originals/ed/2f/9b/ed2f9b6e66b9efefa84d1ee423c718f0.png",
		SupsCost:            "0",
	},
}

var FactionSpecificAbilities = []*server.GameAbility{
	{
		Label:               "AIRSTRIKE",
		FactionID:           FactionIDRedMountain,
		GameClientAbilityID: 6,
		Colour:              "#428EC1",
		ImageUrl:            "https://i.pinimg.com/originals/b1/92/4d/b1924dce177345b5485bb5490ab3441f.jpg",
		SupsCost:            "100000000000000000000",
	},
	{
		Label:               "NUKE",
		FactionID:           FactionIDRedMountain,
		GameClientAbilityID: 7,
		Colour:              "#C24242",
		ImageUrl:            "https://images2.minutemediacdn.com/image/upload/c_crop,h_1126,w_2000,x_0,y_83/f_auto,q_auto,w_1100/v1555949079/shape/mentalfloss/581049-mesut_zengin-istock-1138195821.jpg",
		SupsCost:            "100000000000000000000",
	},
	{
		Label:               "HEAL",
		FactionID:           FactionIDRedMountain,
		GameClientAbilityID: 8,
		Colour:              "#30B07D",
		ImageUrl:            "https://i.pinimg.com/originals/ed/2f/9b/ed2f9b6e66b9efefa84d1ee423c718f0.png",
		SupsCost:            "100000000000000000000",
	},
}

func factionAbilities(ctx context.Context, conn *pgxpool.Pool) error {
	for _, battleAbility := range SharedAbilityCollections {
		err := db.BattleAbilityCreate(ctx, conn, battleAbility)
		if err != nil {
			return terror.Error(err)
		}
	}

	gameclientID := 0
	for _, ability := range SharedFactionAbilities {
		for _, battleAbility := range SharedAbilityCollections {
			if battleAbility.Label == ability.Label {
				ability.BattleAbilityID = &battleAbility.ID
			}
		}
		err := db.GameAbilityCreate(ctx, conn, ability)
		if err != nil {
			return terror.Error(err)
		}
		gameclientID += 1
	}

	// insert red mountain faction abilities
	for _, gameAbility := range FactionSpecificAbilities {
		gameAbility.FactionID = server.RedMountainFactionID
		gameAbility.GameClientAbilityID = byte(gameclientID)

		err := db.GameAbilityCreate(ctx, conn, gameAbility)
		if err != nil {
			return terror.Error(err)
		}

		gameclientID += 1
	}

	// insert boston faction abilities
	for _, gameAbility := range FactionSpecificAbilities {
		gameAbility.FactionID = server.BostonCyberneticsFactionID
		gameAbility.GameClientAbilityID = byte(gameclientID)

		err := db.GameAbilityCreate(ctx, conn, gameAbility)
		if err != nil {
			return terror.Error(err)
		}

		gameclientID += 1
	}

	return nil
}

var factions = []*server.Faction{
	{
		ID:        server.RedMountainFactionID,
		VotePrice: "1000000000000000000",
	},
	{
		ID:        server.BostonCyberneticsFactionID,
		VotePrice: "1000000000000000000",
	},
	{
		ID:        server.ZaibatsuFactionID,
		VotePrice: "1000000000000000000",
	},
}

func (s *Seeder) factions(ctx context.Context) ([]*server.Faction, error) {
	for _, faction := range factions {
		err := db.FactionCreate(ctx, s.Conn, faction)
		if err != nil {
			return nil, terror.Error(err)
		}
	}

	err := db.FactionStatMaterialisedViewRefresh(ctx, s.Conn)
	if err != nil {
		return nil, terror.Error(err)
	}
	return factions, nil
}
