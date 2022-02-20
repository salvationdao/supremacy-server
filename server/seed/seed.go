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

	fmt.Println("Seed streams")
	fmt.Println("Seed streams")
	fmt.Println("Seed streams")
	fmt.Println("Seed streams")
	fmt.Println("Seed streams")
	fmt.Println("Seed streams")
	fmt.Println("Seed streams")
	fmt.Println("Seed streams")
	fmt.Println("Seed streams")

	_, err = s.streams(ctx)
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
		Label:                  "HEAL",
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
		Label:               "HEAL",
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
		Label:               "HEAL",
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
		Label:               "HEAL",
		FactionID:           FactionIDRedMountain,
		GameClientAbilityID: 8,
		Colour:              "#30B07D",
		ImageUrl:            "https://i.pinimg.com/originals/ed/2f/9b/ed2f9b6e66b9efefa84d1ee423c718f0.png",
		SupsCost:            "0",
	},
}

var BostonUniqueAbilities = []*server.GameAbility{
	{

		Label:               "ROBOT DOGS",
		FactionID:           FactionIDBoston,
		GameClientAbilityID: 9,
		Colour:              "#428EC1",
		ImageUrl:            "https://i.pinimg.com/originals/b1/92/4d/b1924dce177345b5485bb5490ab3441f.jpg",
		SupsCost:            "100000000000000000000",
	},
}

var RedMountainUniqueAbilities = []*server.GameAbility{
	{
		Label:               "WAR MACHING REINFORCEMENT",
		FactionID:           FactionIDRedMountain,
		GameClientAbilityID: 10,
		Colour:              "#C24242",
		ImageUrl:            "https://images2.minutemediacdn.com/image/upload/c_crop,h_1126,w_2000,x_0,y_83/f_auto,q_auto,w_1100/v1555949079/shape/mentalfloss/581049-mesut_zengin-istock-1138195821.jpg",
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
	for _, gameAbility := range RedMountainUniqueAbilities {
		err := db.GameAbilityCreate(ctx, conn, gameAbility)
		if err != nil {
			return terror.Error(err)
		}
	}

	// insert boston faction abilities
	for _, gameAbility := range BostonUniqueAbilities {
		err := db.GameAbilityCreate(ctx, conn, gameAbility)
		if err != nil {
			return terror.Error(err)
		}
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

var streams = []*server.Stream{
	{
		Host:          "staging-watch-syd02.supremacy.game",
		Name:          "AU Sydney",
		URL:           "wss://staging-watch-syd02.supremacy.game/WebRTCAppEE/websocket",
		StreamID:      "886200805704583109786601",
		Region:        "au-east",
		Resolution:    "1920x1080",
		BitRatesKBits: 5000,
		UserMax:       1000,
		UsersNow:      100,
		Active:        true,
		Status:        "online",
		Latitude:      -33.9032,
		Longitude:     151.1518,
	},
	{
		Host:          "watch-us-west-1.supremacy.game",
		Name:          "USA Los Angeles",
		URL:           "wss://watch-us-west-1.supremacy.game/WebRTCAppEE/websocket",
		StreamID:      "886200805704583109786601",
		Region:        "us-west",
		Resolution:    "1920x1080",
		BitRatesKBits: 2000,
		UserMax:       1000,
		UsersNow:      370,
		Active:        true,
		Status:        "online",
		Latitude:      34.0522,
		Longitude:     -118.2437,
	},
	{
		Host:          "watch-us-mid-west-1.supremacy.game",
		Name:          "USA Phoenix",
		URL:           "wss://watch-us-mid-west-1.supremacy.game/WebRTCAppEE/websocket",
		StreamID:      "886200805704583109786601",
		Region:        "us-mid-west",
		Resolution:    "1920x1080",
		BitRatesKBits: 1500,
		UserMax:       800,
		UsersNow:      170,
		Active:        true,
		Status:        "online",
		Latitude:      33.6020,
		Longitude:     -111.8879,
	},
	{
		Host:          "watch-au-east-1.supremacy.game",
		Name:          "UK London",
		URL:           "wss://watch-au-east-1.supremacy.game/WebRTCAppEE/websocket",
		StreamID:      "079583650308221367643",
		Region:        "uk",
		Resolution:    "1920x1080",
		BitRatesKBits: 2000,
		UserMax:       1000,
		UsersNow:      200,
		Active:        true,
		Status:        "online",
		Latitude:      51.5085,
		Longitude:     -0.1257,
	},
	{
		Host:          "watch-au-south-1.supremacy.game",
		Name:          "AU Melbourne",
		URL:           "wss://watch-au-south-1.supremacy.game/WebRTCAppEE/websocket",
		StreamID:      "079583650308221367643",
		Region:        "au-south",
		Resolution:    "1920x1080",
		BitRatesKBits: 5000,
		UserMax:       200,
		UsersNow:      100,
		Active:        true,
		Status:        "online",
		Latitude:      -37.8159,
		Longitude:     144.9669,
	},
	{
		Host:          "watch-us-east-1.supremacy.game",
		Name:          "USA New York",
		URL:           "wss://watch-us-east-1.supremacy.game/WebRTCAppEE/websocket",
		StreamID:      "079583650308221367643",
		Region:        "us-east",
		Resolution:    "1920x1080",
		BitRatesKBits: 5000,
		UserMax:       1200,
		UsersNow:      1000,
		Active:        true,
		Status:        "online",
		Latitude:      40.7143,
		Longitude:     -74.0060,
	},
	{
		Host:          "watch-us-nw-1.supremacy.game",
		Name:          "USA Washington",
		URL:           "wss://watch-us-nw-1.supremacy.game/WebRTCAppEE/websocket",
		StreamID:      "079583650308221367643",
		Region:        "us-northwest",
		Resolution:    "1920x1080",
		BitRatesKBits: 5000,
		UserMax:       100,
		UsersNow:      80,
		Active:        true,
		Status:        "online",
		Latitude:      -33.9032,
		Longitude:     151.1518,
	},
	{
		Host:          "watch-au-west-1.supremacy.game",
		Name:          "AU Perth",
		URL:           "wss://watch-au-west-1.supremacy.game/WebRTCAppEE/websocket",
		StreamID:      "079583650308221367643",
		Region:        "au-west",
		Resolution:    "1920x1080",
		BitRatesKBits: 5000,
		UserMax:       100,
		UsersNow:      80,
		Active:        true,
		Status:        "online",
		Latitude:      -31.95000076,
		Longitude:     115.86000061,
	},
	{
		Host:          "watch-eu-1.supremacy.game",
		Name:          "Spain Madrid",
		URL:           "wss://watch-eu-1.supremacy.game/WebRTCAppEE/websocket",
		StreamID:      "079583650308221367643",
		Region:        "eu-spain",
		Resolution:    "1920x1080",
		BitRatesKBits: 900,
		UserMax:       5000,
		UsersNow:      1200,
		Active:        true,
		Status:        "offline",
		Latitude:      40.4165,
		Longitude:     -3.7026,
	},
}

func (s *Seeder) streams(ctx context.Context) ([]*server.Stream, error) {
	for _, stream := range streams {
		err := db.CreateStream(ctx, s.Conn, stream)
		if err != nil {
			return nil, terror.Error(err)
		}
	}

	return streams, nil
}
