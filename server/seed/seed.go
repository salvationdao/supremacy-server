package main

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/ninja-software/terror/v2"
	"io/ioutil"
	"os"
	"time"

	"github.com/georgysavva/scany/pgxscan"

	"github.com/gofrs/uuid"
	"github.com/h2non/filetype"
	"github.com/jackc/pgx/v4/pgxpool"
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
		return err
	}

	fmt.Println("Seed assets")
	_, err = s.assets(ctx)
	if err != nil {
		return err
	}

	fmt.Println("Seed streams")
	_, err = s.streams(ctx)
	if err != nil {
		return err
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
		return err
	}

	return nil
}

func gameMaps(ctx context.Context, conn *pgxpool.Pool) error {
	for _, gameMap := range GameMaps {
		err := GameMapCreate(ctx, conn, gameMap)
		if err != nil {
			return err
		}
	}
	return nil
}

type BattleAbility struct {
	ID                     uuid.UUID `json:"id" db:"id"`
	Label                  string    `json:"label" db:"label"`
	Description            string    `json:"description" db:"description"`
	CooldownDurationSecond int       `json:"cooldown_duration_second" db:"cooldown_duration_second"`
	Colour                 string    `json:"colour"`
	TextColour             string    `json:"text_colour"`
	ImageUrl               string    `json:"image_url"`
}

var BlobIDAbilityAirstrike = uuid.Must(uuid.FromString("dc713e47-4119-494a-a81b-8ac92cf3222b"))
var BlobIDAbilityRobotDogs = uuid.Must(uuid.FromString("3b4ae24a-7ccb-4d3b-8d88-905b406da0e1"))
var BlobIDAbilityReinforcements = uuid.Must(uuid.FromString("5d0a0028-c074-4ab5-b46e-14d0ff07795d"))
var BlobIDAbilityRepair = uuid.Must(uuid.FromString("f40e90b7-1ea2-4a91-bf0f-feb052a019be"))
var BlobIDAbilityNuke = uuid.Must(uuid.FromString("8e0e1918-556c-4370-85f9-b8960fd19554"))
var BlobIDAbilityOvercharge = uuid.Must(uuid.FromString("04acaffd-7bd1-4b01-b264-feb4f8ab4563"))

var FactionIDRedMountain = uuid.Must(uuid.FromString("98bf7bb3-1a7c-4f21-8843-458d62884060"))
var FactionIDBoston = uuid.Must(uuid.FromString("7c6dde21-b067-46cf-9e56-155c88a520e2"))
var FactionIDZaibatsu = uuid.Must(uuid.FromString("880db344-e405-428d-84e5-6ebebab1fe6d"))

var SharedAbilityCollections = []*BattleAbility{
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
		Description:            "Support your Syndicate with a well-timed repair.",
		CooldownDurationSecond: 15,
	},
}

// Blob is a single attachment item on the platform
type Blob struct {
	ID            uuid.UUID  `json:"id" db:"id"`
	FileName      string     `json:"file_name" db:"file_name"`
	MimeType      string     `json:"mime_type" db:"mime_type"`
	FileSizeBytes int64      `json:"file_size_bytes" db:"file_size_bytes"`
	Extension     string     `json:"extension" db:"extension"`
	File          []byte     `json:"file" db:"file"`
	Views         int        `json:"views" db:"views"`
	Hash          *string    `json:"hash" db:"hash"`
	DeletedAt     *time.Time `json:"deleted_at" db:"deleted_at"`
	UpdateAt      *time.Time `json:"updated_at" db:"updated_at"`
	CreatedAt     *time.Time `json:"created_at" db:"created_at"`
}

var AbilityBlobs = []*Blob{
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
	// BlobIDAbilityOvercharge
	{
		ID:       BlobIDAbilityOvercharge,
		FileName: "Overcharge.png",
	},
}

type GameAbility struct {
	ID                  uuid.UUID  `json:"id" db:"id"`
	Identity            uuid.UUID  `json:"identity"` // used for tracking ability price
	GameClientAbilityID byte       `json:"game_client_ability_id" db:"game_client_ability_id"`
	BattleAbilityID     *uuid.UUID `json:"battle_ability_id,omitempty" db:"battle_ability_id,omitempty"`
	Colour              string     `json:"colour" db:"colour"`
	TextColour          string     `json:"text_colour" db:"text_colour"`
	Description         string     `json:"description" db:"description"`
	ImageUrl            string     `json:"image_url" db:"image_url"`
	FactionID           uuid.UUID  `json:"faction_id" db:"faction_id"`
	Label               string     `json:"label" db:"label"`
	SupsCost            string     `json:"sups_cost" db:"sups_cost"`
	CurrentSups         string     `json:"current_sups"`

	// if token id is not 0, it is a nft ability, otherwise it is a faction wide ability
	AbilityHash    string
	WarMachineHash string
	ParticipantID  *byte

	// Category title for frontend to group the abilities together
	Title string `json:"title"`
}

var SharedFactionAbilities = []*GameAbility{
	// FactionIDZaibatsu
	{
		Label:               "AIRSTRIKE",
		FactionID:           FactionIDZaibatsu,
		GameClientAbilityID: 0,
		Colour:              "#173DD1",
		TextColour:          "#FFFFFF",
		Description:         "'Rain fury on the arena with a targeted airstrike.",
		ImageUrl:            "/api/blobs/dc713e47-4119-494a-a81b-8ac92cf3222b",
		SupsCost:            "1000000000000000000",
	},
	{
		Label:               "NUKE",
		FactionID:           FactionIDZaibatsu,
		GameClientAbilityID: 1,
		Colour:              "#E86621",
		TextColour:          "#FFFFFF",
		Description:         "The show-stopper. A tactical nuke at your fingertips.",
		ImageUrl:            "/api/blobs/8e0e1918-556c-4370-85f9-b8960fd19554",
		SupsCost:            "1000000000000000000",
	},
	{
		Label:               "REPAIR",
		FactionID:           FactionIDZaibatsu,
		GameClientAbilityID: 2,
		Colour:              "#23AE3C",
		TextColour:          "#000000",
		Description:         "Support your Syndcate with a well-timed repair.",
		ImageUrl:            "/api/blobs/f40e90b7-1ea2-4a91-bf0f-feb052a019be",
		SupsCost:            "1000000000000000000",
	},
	// FactionIDBoston
	{
		Label:               "AIRSTRIKE",
		GameClientAbilityID: 3,
		FactionID:           FactionIDBoston,
		Colour:              "#173DD1",
		TextColour:          "#FFFFFF",
		Description:         "'Rain fury on the arena with a targeted airstrike.",
		ImageUrl:            "/api/blobs/dc713e47-4119-494a-a81b-8ac92cf3222b",
		SupsCost:            "1000000000000000000",
	},
	{
		Label:               "NUKE",
		FactionID:           FactionIDBoston,
		GameClientAbilityID: 4,
		Colour:              "#E86621",
		TextColour:          "#FFFFFF",
		Description:         "The show-stopper. A tactical nuke at your fingertips.",
		ImageUrl:            "/api/blobs/8e0e1918-556c-4370-85f9-b8960fd19554",
		SupsCost:            "1000000000000000000",
	},
	{
		Label:               "REPAIR",
		FactionID:           FactionIDBoston,
		GameClientAbilityID: 5,
		Colour:              "#23AE3C",
		TextColour:          "#000000",
		Description:         "Support your Syndcate with a well-timed repair.",
		ImageUrl:            "/api/blobs/f40e90b7-1ea2-4a91-bf0f-feb052a019be",
		SupsCost:            "1000000000000000000",
	},
	// FactionIDRedMountain
	{
		Label:               "AIRSTRIKE",
		FactionID:           FactionIDRedMountain,
		GameClientAbilityID: 6,
		Colour:              "#173DD1",
		TextColour:          "#FFFFFF",
		Description:         "'Rain fury on the arena with a targeted airstrike.",
		ImageUrl:            "/api/blobs/dc713e47-4119-494a-a81b-8ac92cf3222b",
		SupsCost:            "1000000000000000000",
	},
	{
		Label:               "NUKE",
		FactionID:           FactionIDRedMountain,
		GameClientAbilityID: 7,
		Colour:              "#E86621",
		TextColour:          "#FFFFFF",
		Description:         "The show-stopper. A tactical nuke at your fingertips.",
		ImageUrl:            "/api/blobs/8e0e1918-556c-4370-85f9-b8960fd19554",
		SupsCost:            "1000000000000000000",
	},
	{
		Label:               "REPAIR",
		FactionID:           FactionIDRedMountain,
		GameClientAbilityID: 8,
		Colour:              "#23AE3C",
		TextColour:          "#000000",
		Description:         "Support your Syndcate with a well-timed repair.",
		ImageUrl:            "/api/blobs/f40e90b7-1ea2-4a91-bf0f-feb052a019be",
		SupsCost:            "1000000000000000000",
	},
}

var BostonUniqueAbilities = []*GameAbility{
	{

		Label:               "ROBOT DOGS",
		FactionID:           FactionIDBoston,
		GameClientAbilityID: 9,
		Colour:              "#428EC1",
		TextColour:          "#FFFFFF",
		Description:         "Boston Cybernetic unique ability. Release the hounds!",
		ImageUrl:            "/api/blobs/3b4ae24a-7ccb-4d3b-8d88-905b406da0e1",
		SupsCost:            "100000000000000000000",
	},
}

var RedMountainUniqueAbilities = []*GameAbility{
	{
		Label:               "REINFORCEMENTS",
		FactionID:           FactionIDRedMountain,
		GameClientAbilityID: 10,
		Colour:              "#C52A1F",
		TextColour:          "#FFFFFF",
		Description:         "Red Mountain unique ability. Call an additional Mech to the arena.",
		ImageUrl:            "/api/blobs/5d0a0028-c074-4ab5-b46e-14d0ff07795d",
		SupsCost:            "100000000000000000000",
	},
}

var ZaibatsuUniqueAbilities = []*GameAbility{
	{
		Label:               "OVERCHARGE",
		FactionID:           FactionIDZaibatsu,
		GameClientAbilityID: 11,
		Colour:              "#FFFFFF",
		TextColour:          "#000000",
		Description:         "Zaibatsu unique ability. Consume your remaining shield for an explosive defence mechanism.",
		ImageUrl:            "/api/blobs/04acaffd-7bd1-4b01-b264-feb4f8ab4563",
		SupsCost:            "100000000000000000000",
	},
}

// GameAbilityCreate create a new faction action
func GameAbilityCreate(ctx context.Context, conn *pgxpool.Pool, gameAbility *GameAbility) error {
	q := `
		INSERT INTO
			game_abilities (game_client_ability_id, faction_id, label, sups_cost, battle_ability_id, colour, text_colour, description, image_url)
		VALUES
			($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING
			id, game_client_ability_id, faction_id, label, sups_cost, battle_ability_id, colour, text_colour, description, image_url
	`

	err := pgxscan.Get(ctx, conn, gameAbility, q,
		gameAbility.GameClientAbilityID,
		gameAbility.FactionID,
		gameAbility.Label,
		gameAbility.SupsCost,
		gameAbility.BattleAbilityID,
		gameAbility.Colour,
		gameAbility.TextColour,
		gameAbility.Description,
		gameAbility.ImageUrl,
	)
	if err != nil {
		return err
	}

	return nil
}

// BattleAbilityCreate create ability collection
func BattleAbilityCreate(ctx context.Context, conn *pgxpool.Pool, battleAbility *BattleAbility) error {
	q := `
		INSERT INTO
			battle_abilities (label, description, cooldown_duration_second)
		VALUES
			($1, $2, $3)
		RETURNING
			id, label, description, cooldown_duration_second
	`
	err := pgxscan.Get(ctx, conn, battleAbility, q,
		battleAbility.Label,
		battleAbility.Description,
		battleAbility.CooldownDurationSecond,
	)
	if err != nil {
		return err
	}

	return nil
}

func factionAbilities(ctx context.Context, conn *pgxpool.Pool) error {
	for _, battleAbility := range SharedAbilityCollections {
		err := BattleAbilityCreate(ctx, conn, battleAbility)
		if err != nil {
			return err
		}
	}

	gameclientID := 0
	for _, ability := range SharedFactionAbilities {
		for _, battleAbility := range SharedAbilityCollections {
			if battleAbility.Label == ability.Label {
				ability.BattleAbilityID = &battleAbility.ID
			}
		}
		err := GameAbilityCreate(ctx, conn, ability)
		if err != nil {
			return err
		}
		gameclientID += 1
	}

	// insert red mountain faction abilities
	for _, gameAbility := range RedMountainUniqueAbilities {
		err := GameAbilityCreate(ctx, conn, gameAbility)
		if err != nil {
			return err
		}
	}

	// insert boston faction abilities
	for _, gameAbility := range BostonUniqueAbilities {
		err := GameAbilityCreate(ctx, conn, gameAbility)
		if err != nil {
			return err
		}
	}

	// insert zaibatsu faction abilities
	for _, gameAbility := range ZaibatsuUniqueAbilities {
		err := GameAbilityCreate(ctx, conn, gameAbility)
		if err != nil {
			return err
		}
	}

	return nil
}

type Faction struct {
	ID               uuid.UUID     `json:"id" db:"id"`
	Label            string        `json:"label" db:"label"`
	Theme            *FactionTheme `json:"theme" db:"theme"`
	LogoBlobID       uuid.UUID     `json:"logo_blob_id,omitempty"`
	BackgroundBlobID uuid.UUID     `json:"background_blob_id,omitempty"`
	VotePrice        string        `json:"vote_price" db:"vote_price"`
	ContractReward   string        `json:"contract_reward" db:"contract_reward"`
}

type FactionTheme struct {
	Primary    string `json:"primary"`
	Secondary  string `json:"secondary"`
	Background string `json:"background"`
}

var RedMountainFactionID = uuid.Must(uuid.FromString("98bf7bb3-1a7c-4f21-8843-458d62884060"))
var BostonCyberneticsFactionID = uuid.Must(uuid.FromString("7c6dde21-b067-46cf-9e56-155c88a520e2"))
var ZaibatsuFactionID = uuid.Must(uuid.FromString("880db344-e405-428d-84e5-6ebebab1fe6d"))

var factions = []*Faction{
	{
		ID:        RedMountainFactionID,
		VotePrice: "1000000000000000000",
	},
	{
		ID:        BostonCyberneticsFactionID,
		VotePrice: "1000000000000000000",
	},
	{
		ID:        ZaibatsuFactionID,
		VotePrice: "1000000000000000000",
	},
}

// FactionCreate create a new faction
func FactionCreate(ctx context.Context, conn *pgxpool.Pool, faction *Faction) error {
	q := `
		INSERT INTO
			factions (id, vote_price)
		VALUES
			($1, $2)
		RETURNING
			id, vote_price
	`

	err := pgxscan.Get(ctx, conn, faction, q, faction.ID, faction.VotePrice)
	if err != nil {
		return err
	}

	return nil
}

// FactionStatMaterialisedViewRefresh
func FactionStatMaterialisedViewRefresh(ctx context.Context, conn *pgxpool.Pool) error {
	q := `
		REFRESH MATERIALIZED VIEW faction_stats;
	`
	_, err := conn.Exec(ctx, q)
	if err != nil {
		return err
	}

	return nil
}

func (s *Seeder) factions(ctx context.Context) ([]*Faction, error) {
	for _, faction := range factions {
		err := FactionCreate(ctx, s.Conn, faction)
		if err != nil {
			return nil, err
		}
	}

	err := FactionStatMaterialisedViewRefresh(ctx, s.Conn)
	if err != nil {
		return nil, err
	}
	return factions, nil
}

type Stream struct {
	Host          string  `json:"host" db:"host"`
	Name          string  `json:"name" db:"name"`
	StreamID      string  `json:"streamID" db:"stream_id"`
	URL           string  `json:"url" db:"url"`
	Region        string  `json:"region" db:"region"`
	Resolution    string  `json:"resolution" db:"resolution"`
	BitRatesKBits int     `json:"bitRatesKBits" db:"bit_rates_k_bits"`
	UserMax       int     `json:"userMax" db:"user_max"`
	UsersNow      int     `json:"usersNow" db:"users_now"`
	Active        bool    `json:"active" db:"active"`
	Status        string  `json:"status" db:"status"`
	Latitude      float32 `json:"latitude" db:"latitude"`
	Longitude     float32 `json:"longitude" db:"longitude"`
}

var streams = []*Stream{

	// singapore
	{
		Host:          "https://video-sg.ninja-cdn.com/WebRTCAppEE/player.html?name=R3dvaIhZOxRr1645381571194",
		Name:          "Singapore",
		URL:           "wss://video-sg.ninja-cdn.com/WebRTCAppEE/websocket",
		StreamID:      "R3dvaIhZOxRr1645381571194",
		Region:        "se-asia",
		Resolution:    "1920x1080",
		BitRatesKBits: 4000,
		UserMax:       1000,
		UsersNow:      100,
		Active:        true,
		Status:        "online",
		Latitude:      1.3521,
		Longitude:     103.8198,
	},

	// Germany
	{
		Host:          "https://video-de.ninja-cdn.com/WebRTCAppEE/player.html?name=R3dvaIhZOxRr1645381571194",
		Name:          "Germany",
		URL:           "wss://video-de.ninja-cdn.com/WebRTCAppEE/websocket",
		StreamID:      "R3dvaIhZOxRr1645381571194",
		Region:        "eu",
		Resolution:    "1920x1080",
		BitRatesKBits: 4000,
		UserMax:       1000,
		UsersNow:      100,
		Active:        true,
		Status:        "online",
		Latitude:      10.4515,
		Longitude:     51.1657,
	},

	// USA
	{
		Host:          "https://video.ninja-cdn.com/WebRTCAppEE/player.html?name=R3dvaIhZOxRr1645381571194",
		Name:          "USA",
		URL:           "wss://video.ninja-cdn.com/WebRTCAppEE/websocket",
		StreamID:      "R3dvaIhZOxRr1645381571194",
		Region:        "us",
		Resolution:    "1920x1080",
		BitRatesKBits: 4000,
		UserMax:       1000,
		UsersNow:      100,
		Active:        true,
		Status:        "online",
		Latitude:      95.7129,
		Longitude:     37.0902,
	},
}

// CreateStream created a new stream
func CreateStream(ctx context.Context, conn *pgxpool.Pool, stream *Stream) error {
	q := `
		INSERT INTO
			stream_list (host, name, url, stream_id, region, resolution, bit_rates_k_bits, user_max, users_now, active, status, latitude, longitude)
		VALUES
			($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING
		host, name, url, stream_id region, resolution, bit_rates_k_bits, user_max, users_now, active, status
	`

	err := pgxscan.Get(ctx, conn, stream, q, stream.Host, stream.Name, stream.URL, stream.StreamID, stream.Region, stream.Resolution, stream.BitRatesKBits, stream.UserMax, stream.UsersNow, stream.Active, stream.Status, stream.Latitude, stream.Longitude)
	if err != nil {
		return err
	}

	return nil
}

func (s *Seeder) streams(ctx context.Context) ([]*Stream, error) {
	for _, stream := range streams {
		err := CreateStream(ctx, s.Conn, stream)
		if err != nil {
			return nil, err
		}
	}

	return streams, nil
}

func (s *Seeder) assets(ctx context.Context) ([]*Blob, error) {
	output := []*Blob{}
	for _, blob := range AbilityBlobs {
		f, err := os.Open("./asset/" + blob.FileName)
		if err != nil {
			return nil, err
		}
		fileData, err := ioutil.ReadAll(f)
		if err != nil {
			return nil, err
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

		err = BlobInsert(ctx, s.Conn, blob, blob.ID, blob.FileName, blob.MimeType, blob.FileSizeBytes, blob.Extension, blob.File, blob.Hash)
		if err != nil {
			return nil, terror.Error(err, "blob upsert error")
		}

		output = append(output, blob)
	}
	return output, nil
}

// BlobInsert inserts a new blob
func BlobInsert(ctx context.Context, conn *pgxpool.Pool, result *Blob, id uuid.UUID, fileName string, mimeType string, fileSizeBytes int64, extension string, file []byte, hash *string) error {
	q := `
		INSERT INTO blobs (id, file_name, mime_type, file_size_bytes, extension, file, hash) 
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (id) DO UPDATE
		SET file_name = EXCLUDED.file_name,
			mime_type = EXCLUDED.mime_type,
			file_size_bytes = EXCLUDED.file_size_bytes,
			extension = EXCLUDED.extension,
			file = EXCLUDED.file,
			hash = EXCLUDED.hash
		RETURNING id, file_name, mime_type, file_size_bytes, extension, file, hash`
	err := pgxscan.Get(ctx, conn, result, q, id, fileName, mimeType, fileSizeBytes, extension, file, hash)
	if err != nil {
		return err
	}
	return nil
}
