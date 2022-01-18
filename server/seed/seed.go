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

// MaxMembersPerOrganisation is the default amount of member users per organisation (also includes non-organisation users)
const MaxMembersPerOrganisation = 5

// MaxTestUsers is the default amount of user for the test organisation account (first organisation has X reserved test users)
const MaxTestUsers = 3

type Seeder struct {
	Conn *pgxpool.Pool
}

// NewSeeder returns a new Seeder
func NewSeeder(conn *pgxpool.Pool) *Seeder {
	s := &Seeder{conn}
	return s
}

// Run for database spinup
func (s *Seeder) Run() error {
	// seed review icons
	fmt.Println("Seed game maps")
	ctx := context.Background()
	err := gameMaps(ctx, s.Conn)
	if err != nil {
		return terror.Error(err)
	}

	// fmt.Println("Seed factions")
	// err = factions(ctx, s.Conn)
	// if err != nil {
	// 	return terror.Error(err)
	// }

	fmt.Println("Seed faction abilities")
	err = factionAbilities(ctx, s.Conn)
	if err != nil {
		return terror.Error(err)
	}

	// fmt.Println("Seed war machines")
	// err = warMachines(ctx, s.Conn)
	// if err != nil {
	// 	return terror.Error(err)
	// }

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

// var Factions = []*server.Faction{
// 	{
// 		ID:     server.FactionID(uuid.Must(uuid.FromString("98bf7bb3-1a7c-4f21-8843-458d62884060"))),
// 		Label:  "Red Mountain Offworld Mining Corporation",
// 		Colour: "#BB1C2A",
// 	},
// 	{
// 		ID:     server.FactionID(uuid.Must(uuid.FromString("7c6dde21-b067-46cf-9e56-155c88a520e2"))),
// 		Label:  "Boston Cybernetics",
// 		Colour: "#03AAF9",
// 	},
// 	{
// 		ID:     server.FactionID(uuid.Must(uuid.FromString("880db344-e405-428d-84e5-6ebebab1fe6d"))),
// 		Label:  "Zaibatsu Heavy Industries",
// 		Colour: "#263D4D",
// 	},
// }

// func factions(ctx context.Context, conn *pgxpool.Pool) error {
// 	for _, faction := range Factions {
// 		err := db.FactionCreate(ctx, conn, faction)
// 		if err != nil {
// 			return terror.Error(err)
// 		}
// 	}
// 	return nil
// }

var factionIDs = []server.FactionID{
	server.FactionID(uuid.Must(uuid.FromString("98bf7bb3-1a7c-4f21-8843-458d62884060"))),
	server.FactionID(uuid.Must(uuid.FromString("7c6dde21-b067-46cf-9e56-155c88a520e2"))),
	server.FactionID(uuid.Must(uuid.FromString("880db344-e405-428d-84e5-6ebebab1fe6d"))),
}

var FactionAbilities = []*server.FactionAbility{
	{
		Label:                  "AIRSTRIKE",
		Type:                   server.FactionAbilityTypeAirStrike,
		Colour:                 "#428EC1",
		SupsCost:               60,
		ImageUrl:               "https://i.pinimg.com/originals/b1/92/4d/b1924dce177345b5485bb5490ab3441f.jpg",
		CooldownDurationSecond: 15,
	},
	{
		Label:                  "NUKE",
		Type:                   server.FactionAbilityTypeNuke,
		Colour:                 "#C24242",
		SupsCost:               60,
		ImageUrl:               "https://images2.minutemediacdn.com/image/upload/c_crop,h_1126,w_2000,x_0,y_83/f_auto,q_auto,w_1100/v1555949079/shape/mentalfloss/581049-mesut_zengin-istock-1138195821.jpg",
		CooldownDurationSecond: 20,
	},
	{
		Label:                  "HEAL",
		Type:                   server.FactionAbilityTypeHealing,
		Colour:                 "#30B07D",
		SupsCost:               60,
		ImageUrl:               "https://i.pinimg.com/originals/ed/2f/9b/ed2f9b6e66b9efefa84d1ee423c718f0.png",
		CooldownDurationSecond: 10,
	},
}

func factionAbilities(ctx context.Context, conn *pgxpool.Pool) error {
	for _, factionID := range factionIDs {
		for _, ability := range FactionAbilities {
			ability.FactionID = factionID
			err := db.FactionAbilityCreate(ctx, conn, ability)
			if err != nil {
				return terror.Error(err)
			}
		}
	}
	return nil
}

// var WarMachines = []*server.WarMachine{
// 	{
// 		Name:            "Zeus",
// 		BaseHealthPoint: 100,
// 		BaseShieldPoint: 120,
// 	},
// 	{
// 		Name:            "Poseidon",
// 		BaseHealthPoint: 100,
// 		BaseShieldPoint: 120,
// 	},
// 	{
// 		Name:            "Hera",
// 		BaseHealthPoint: 100,
// 		BaseShieldPoint: 120,
// 	},
// 	{
// 		Name:            "Athena",
// 		BaseHealthPoint: 100,
// 		BaseShieldPoint: 120,
// 	},
// 	{
// 		Name:            "Hercules",
// 		BaseHealthPoint: 100,
// 		BaseShieldPoint: 120,
// 	},
// 	{
// 		Name:            "Hephaestus",
// 		BaseHealthPoint: 100,
// 		BaseShieldPoint: 120,
// 	},
// }

// func warMachines(ctx context.Context, conn *pgxpool.Pool) error {
// 	for i, warMachine := range WarMachines {
// 		warMachine.ID = uint64(i)
// 		err := db.WarMachineCreate(ctx, conn, warMachine)
// 		if err != nil {
// 			return terror.Error(err)
// 		}
// 	}
// 	return nil
// }
