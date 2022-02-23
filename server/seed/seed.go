package seed

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"server"
	"server/db"

	"github.com/gofrs/uuid"
	"github.com/h2non/filetype"
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

	fmt.Println("Seed assets")
	_, err = s.assets(ctx)
	if err != nil {
		return terror.Error(err)
	}

	fmt.Println("Seed faction abilities")
	err = factionAbilities(ctx, s.Conn)
	if err != nil {
		return terror.Error(err)
	}

	fmt.Println("Seed streams")
	_, err = s.streams(ctx)
	if err != nil {
		return terror.Error(err)
	}

	fmt.Println("Seed complete!")

	return nil
}

// Run for database spin up (post prod)
func (s *Seeder) RunAssets() error {
	ctx := context.Background()
	fmt.Println("Seed assets")
	_, err := s.assets(ctx)
	if err != nil {
		return terror.Error(err)
	}
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

var BlobIDAbilityAirstrike = server.BlobID(uuid.Must(uuid.FromString("dc713e47-4119-494a-a81b-8ac92cf3222b")))
var BlobIDAbilityRobotDogs = server.BlobID(uuid.Must(uuid.FromString("3b4ae24a-7ccb-4d3b-8d88-905b406da0e1")))
var BlobIDAbilityReinforcements = server.BlobID(uuid.Must(uuid.FromString("5d0a0028-c074-4ab5-b46e-14d0ff07795d")))
var BlobIDAbilityRepair = server.BlobID(uuid.Must(uuid.FromString("f40e90b7-1ea2-4a91-bf0f-feb052a019be")))
var BlobIDAbilityNuke = server.BlobID(uuid.Must(uuid.FromString("8e0e1918-556c-4370-85f9-b8960fd19554")))

var FactionIDRedMountain = server.FactionID(uuid.Must(uuid.FromString("98bf7bb3-1a7c-4f21-8843-458d62884060")))
var FactionIDBoston = server.FactionID(uuid.Must(uuid.FromString("7c6dde21-b067-46cf-9e56-155c88a520e2")))
var FactionIDZaibatsu = server.FactionID(uuid.Must(uuid.FromString("880db344-e405-428d-84e5-6ebebab1fe6d")))

var SharedAbilityCollections = []*server.BattleAbility{
	{
		Label:                  "AIRSTRIKE",
		Description:            "Rain fury on the arena with a targeted airstrike.",
		CooldownDurationSecond: 20,
	},
	{
		Label:                  "NUKE",
		Description:            "The show-stopper. A tactical nuke at your fingertips.",
		CooldownDurationSecond: 30,
	},
	{
		Label:                  "REPAIR",
		Description:            "Support your Syndcate with a well-timed repair.",
		CooldownDurationSecond: 15,
	},
}

var AbilityBlobs = []*server.Blob{
	// BlobIDAbilityAirstrike
	{
		ID:       BlobIDAbilityAirstrike,
		FileName: "Airstrike.png",
	},
	// BlobIDAbilityRobotDogs
	{
		ID:       BlobIDAbilityRobotDogs,
		FileName: "Dogs.png",
	},
	// BlobIDAbilityReinforacements
	{
		ID:       BlobIDAbilityReinforcements,
		FileName: "Reinforcements.png",
	},
	// BlobIDAbilityRepair
	{
		ID:       BlobIDAbilityRepair,
		FileName: "Repair.png",
	},
	// BlobIDAbilityNuke
	{
		ID:       BlobIDAbilityNuke,
		FileName: "Nuke.png",
	},
}

var SharedFactionAbilities = []*server.GameAbility{
	// FactionIDZaibatsu
	{
		Label:               "AIRSTRIKE",
		FactionID:           FactionIDZaibatsu,
		GameClientAbilityID: 0,
		Colour:              "#3B5DAD",
		Description:         "'Rain fury on the arena with a targeted airstrike.",
		ImageUrl:            "/api/blobs/dc713e47-4119-494a-a81b-8ac92cf3222b",
		SupsCost:            "0",
	},
	{
		Label:               "NUKE",
		FactionID:           FactionIDZaibatsu,
		GameClientAbilityID: 1,
		Colour:              "#B8422A",
		Description:         "The show-stopper. A tactical nuke at your fingertips.",
		ImageUrl:            "/api/blobs/8e0e1918-556c-4370-85f9-b8960fd19554",
		SupsCost:            "0",
	},
	{
		Label:               "REPAIR",
		FactionID:           FactionIDZaibatsu,
		GameClientAbilityID: 2,
		Colour:              "#25A16F",
		Description:         "Support your Syndcate with a well-timed repair.",
		ImageUrl:            "/api/blobs/f40e90b7-1ea2-4a91-bf0f-feb052a019be",
		SupsCost:            "0",
	},
	// FactionIDBoston
	{
		Label:               "AIRSTRIKE",
		GameClientAbilityID: 3,
		FactionID:           FactionIDBoston,
		Colour:              "#3B5DAD",
		Description:         "'Rain fury on the arena with a targeted airstrike.",
		ImageUrl:            "/api/blobs/dc713e47-4119-494a-a81b-8ac92cf3222b",
		SupsCost:            "0",
	},
	{
		Label:               "NUKE",
		FactionID:           FactionIDBoston,
		GameClientAbilityID: 4,
		Colour:              "#B8422A",
		Description:         "The show-stopper. A tactical nuke at your fingertips.",
		ImageUrl:            "/api/blobs/8e0e1918-556c-4370-85f9-b8960fd19554",
		SupsCost:            "0",
	},
	{
		Label:               "REPAIR",
		FactionID:           FactionIDBoston,
		GameClientAbilityID: 5,
		Colour:              "#25A16F",
		Description:         "Support your Syndcate with a well-timed repair.",
		ImageUrl:            "/api/blobs/f40e90b7-1ea2-4a91-bf0f-feb052a019be",
		SupsCost:            "0",
	},
	// FactionIDRedMountain
	{
		Label:               "AIRSTRIKE",
		FactionID:           FactionIDRedMountain,
		GameClientAbilityID: 6,
		Colour:              "#3B5DAD",
		Description:         "'Rain fury on the arena with a targeted airstrike.",
		ImageUrl:            "/api/blobs/dc713e47-4119-494a-a81b-8ac92cf3222b",
		SupsCost:            "0",
	},
	{
		Label:               "NUKE",
		FactionID:           FactionIDRedMountain,
		GameClientAbilityID: 7,
		Colour:              "#B8422A",
		Description:         "The show-stopper. A tactical nuke at your fingertips.",
		ImageUrl:            "/api/blobs/8e0e1918-556c-4370-85f9-b8960fd19554",
		SupsCost:            "0",
	},
	{
		Label:               "REPAIR",
		FactionID:           FactionIDRedMountain,
		GameClientAbilityID: 8,
		Colour:              "#25A16F",
		Description:         "Support your Syndcate with a well-timed repair.",
		ImageUrl:            "/api/blobs/f40e90b7-1ea2-4a91-bf0f-feb052a019be",
		SupsCost:            "0",
	},
}

var BostonUniqueAbilities = []*server.GameAbility{
	{

		Label:               "ROBOT DOGS",
		FactionID:           FactionIDBoston,
		GameClientAbilityID: 9,
		Colour:              "#6F40AD",
		Description:         "Boston Cybernetic unique ability. Release the hounds!",
		ImageUrl:            "/api/blobs/3b4ae24a-7ccb-4d3b-8d88-905b406da0e1",
		SupsCost:            "100000000000000000000",
	},
}

var RedMountainUniqueAbilities = []*server.GameAbility{
	{
		Label:               "REINFORCEMENTS",
		FactionID:           FactionIDRedMountain,
		GameClientAbilityID: 10,
		Colour:              "#C42B40",
		Description:         "Red Mountain unique ability. Call an additional Mech to the arena.",
		ImageUrl:            "/api/blobs/5d0a0028-c074-4ab5-b46e-14d0ff07795d",
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
		Name:          "USA",
		URL:           "wss://video.ninja-cdn.com:5443/WebRTCAppEE/websocket",
		StreamID:      "R3dvaIhZOxRr1645381571194",
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
	// {
	// 	Host:          "watch-us-west-1.supremacy.game",
	// 	Name:          "USA Los Angeles",
	// 	URL:           "wss://watch-us-west-1.supremacy.game/WebRTCAppEE/websocket",
	// 	StreamID:      "886200805704583109786601",
	// 	Region:        "us-west",
	// 	Resolution:    "1920x1080",
	// 	BitRatesKBits: 2000,
	// 	UserMax:       1000,
	// 	UsersNow:      370,
	// 	Active:        true,
	// 	Status:        "online",
	// 	Latitude:      34.0522,
	// 	Longitude:     -118.2437,
	// },
	// {
	// 	Host:          "watch-us-mid-west-1.supremacy.game",
	// 	Name:          "USA Phoenix",
	// 	URL:           "wss://watch-us-mid-west-1.supremacy.game/WebRTCAppEE/websocket",
	// 	StreamID:      "886200805704583109786601",
	// 	Region:        "us-mid-west",
	// 	Resolution:    "1920x1080",
	// 	BitRatesKBits: 1500,
	// 	UserMax:       800,
	// 	UsersNow:      170,
	// 	Active:        true,
	// 	Status:        "online",
	// 	Latitude:      33.6020,
	// 	Longitude:     -111.8879,
	// },
	// {
	// 	Host:          "watch-au-east-1.supremacy.game",
	// 	Name:          "UK London",
	// 	URL:           "wss://watch-au-east-1.supremacy.game/WebRTCAppEE/websocket",
	// 	StreamID:      "079583650308221367643",
	// 	Region:        "uk",
	// 	Resolution:    "1920x1080",
	// 	BitRatesKBits: 2000,
	// 	UserMax:       1000,
	// 	UsersNow:      200,
	// 	Active:        true,
	// 	Status:        "online",
	// 	Latitude:      51.5085,
	// 	Longitude:     -0.1257,
	// },
	// {
	// 	Host:          "watch-au-south-1.supremacy.game",
	// 	Name:          "AU Melbourne",
	// 	URL:           "wss://watch-au-south-1.supremacy.game/WebRTCAppEE/websocket",
	// 	StreamID:      "079583650308221367643",
	// 	Region:        "au-south",
	// 	Resolution:    "1920x1080",
	// 	BitRatesKBits: 5000,
	// 	UserMax:       200,
	// 	UsersNow:      100,
	// 	Active:        true,
	// 	Status:        "online",
	// 	Latitude:      -37.8159,
	// 	Longitude:     144.9669,
	// },
	// {
	// 	Host:          "watch-us-east-1.supremacy.game",
	// 	Name:          "USA New York",
	// 	URL:           "wss://watch-us-east-1.supremacy.game/WebRTCAppEE/websocket",
	// 	StreamID:      "079583650308221367643",
	// 	Region:        "us-east",
	// 	Resolution:    "1920x1080",
	// 	BitRatesKBits: 5000,
	// 	UserMax:       1200,
	// 	UsersNow:      1000,
	// 	Active:        true,
	// 	Status:        "online",
	// 	Latitude:      40.7143,
	// 	Longitude:     -74.0060,
	// },
	// {
	// 	Host:          "watch-us-nw-1.supremacy.game",
	// 	Name:          "USA Washington",
	// 	URL:           "wss://watch-us-nw-1.supremacy.game/WebRTCAppEE/websocket",
	// 	StreamID:      "079583650308221367643",
	// 	Region:        "us-northwest",
	// 	Resolution:    "1920x1080",
	// 	BitRatesKBits: 5000,
	// 	UserMax:       100,
	// 	UsersNow:      80,
	// 	Active:        true,
	// 	Status:        "online",
	// 	Latitude:      -33.9032,
	// 	Longitude:     151.1518,
	// },
	// {
	// 	Host:          "watch-au-west-1.supremacy.game",
	// 	Name:          "AU Perth",
	// 	URL:           "wss://watch-au-west-1.supremacy.game/WebRTCAppEE/websocket",
	// 	StreamID:      "079583650308221367643",
	// 	Region:        "au-west",
	// 	Resolution:    "1920x1080",
	// 	BitRatesKBits: 5000,
	// 	UserMax:       100,
	// 	UsersNow:      80,
	// 	Active:        true,
	// 	Status:        "online",
	// 	Latitude:      -31.95000076,
	// 	Longitude:     115.86000061,
	// },
	// {
	// 	Host:          "watch-eu-1.supremacy.game",
	// 	Name:          "Spain Madrid",
	// 	URL:           "wss://watch-eu-1.supremacy.game/WebRTCAppEE/websocket",
	// 	StreamID:      "079583650308221367643",
	// 	Region:        "eu-spain",
	// 	Resolution:    "1920x1080",
	// 	BitRatesKBits: 900,
	// 	UserMax:       5000,
	// 	UsersNow:      1200,
	// 	Active:        true,
	// 	Status:        "offline",
	// 	Latitude:      40.4165,
	// 	Longitude:     -3.7026,
	// },
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

func (s *Seeder) assets(ctx context.Context) ([]*server.Blob, error) {
	output := []*server.Blob{}
	for _, blob := range AbilityBlobs {
		f, err := os.Open("./asset/" + blob.FileName)
		if err != nil {
			return nil, terror.Error(err)
		}
		fileData, err := ioutil.ReadAll(f)
		if err != nil {
			return nil, terror.Error(err)
		}

		// Get image mime type
		kind, err := filetype.Match(fileData)
		if err != nil {
			return nil, terror.Error(terror.ErrParse, "parse error")
		}

		if kind == filetype.Unknown {
			return nil, terror.Error(fmt.Errorf("Image type is unknown"), "Image type is unknown")
		}

		mimeType := kind.MIME.Value
		extension := kind.Extension

		// Get hash
		hasher := md5.New()
		_, err = hasher.Write(fileData)
		if err != nil {
			return nil, terror.Error(err, "hash error")
		}
		hashResult := hasher.Sum(nil)
		hash := hex.EncodeToString(hashResult)

		blob.MimeType = mimeType
		blob.Extension = extension
		blob.FileSizeBytes = int64(len(fileData))
		blob.File = fileData
		blob.Hash = &hash

		err = db.BlobInsert(ctx, s.Conn, blob, blob.ID, blob.FileName, blob.MimeType, blob.FileSizeBytes, blob.Extension, blob.File, blob.Hash)
		if err != nil {
			return nil, terror.Error(err, "blob insert error")
		}

		output = append(output, blob)
	}
	return output, nil
}
