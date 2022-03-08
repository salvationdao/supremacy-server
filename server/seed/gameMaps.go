package main

import (
	"context"

	"github.com/georgysavva/scany/pgxscan"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/ninja-software/terror/v2"
)

// To get the location in game its
//  ((cellX * GameClientTileSize) + GameClientTileSize / 2) + LeftPixels
//  ((cellY * GameClientTileSize) + GameClientTileSize / 2) + TopPixels

type GameMap struct {
	ID            uuid.UUID `json:"id" db:"id"`
	Name          string    `json:"name" db:"name"`
	ImageUrl      string    `json:"image_url" db:"image_url"`
	MaxSpawns     int       `json:"max_spawns" db:"max_spawns"`
	Width         int       `json:"width" db:"width"`
	Height        int       `json:"height" db:"height"`
	CellsX        int       `json:"cells_x" db:"cells_x"`
	CellsY        int       `json:"cells_y" db:"cells_y"`
	TopPixels     int       `json:"top" db:"top_pixels"`
	LeftPixels    int       `json:"left" db:"left_pixels"`
	Scale         float64   `json:"scale" db:"scale"`
	DisabledCells []int     `json:"disabled_cells" db:"disabled_cells"`
}

var GameMaps = []*GameMap{
	{
		Name:       "DesertCity",
		ImageUrl:   "https://ninjasoftware-static-media.s3.ap-southeast-2.amazonaws.com/supremacy/maps/desert_city.jpg",
		Width:      1700,
		Height:     1600,
		MaxSpawns:  9,
		CellsX:     34,
		CellsY:     32,
		TopPixels:  -40000,
		LeftPixels: -40000,
		Scale:      0.025,
		DisabledCells: []int{
			0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32, 33,
			34, 35, 36, 37, 38, 39, 40, 41, 42, 43, 44, 45, 46, 47, 48, 49, 50, 51, 52, 53, 54, 55, 56, 57, 58, 59, 60, 61, 62, 63, 64, 65, 66, 67,
			68, 69, 70, 71, 72, 73, 74, 75, 76, 77, 78, 79, 80, 81, 82, 83, 84, 85, 86, 87, 88, 89, 90, 91, 92, 93, 94, 95, 96, 97, 98, 99, 100, 101,
			102, 103, 104, 105, 132, 133, 134, 135,
			136, 137, 138, 139, 166, 167, 168, 169,
			170, 171, 172, 173, 200, 201, 202, 203,
			204, 205, 206, 207, 234, 235, 236, 237,
			238, 239, 240, 241, 268, 269, 270, 271,
			272, 273, 274, 275, 302, 303, 304, 305,
			306, 307, 308, 309, 336, 337, 338, 339,
			340, 341, 342, 343, 370, 371, 372, 373,
			374, 375, 376, 377, 404, 405, 406, 407,
			408, 409, 410, 411, 438, 439, 440, 441,
			442, 443, 444, 445, 472, 473, 474, 475,
			476, 477, 478, 479, 506, 507, 508, 509,
			510, 511, 512, 513, 540, 541, 542, 543,
			544, 545, 546, 547, 574, 575, 576, 577,
			578, 579, 580, 581, 608, 609, 610, 611,
			612, 613, 614, 615, 642, 643, 644, 645,
			646, 647, 648, 649, 676, 677, 678, 679,
			680, 681, 682, 683, 710, 711, 712, 713,
			714, 715, 716, 717, 744, 745, 746, 747,
			748, 749, 750, 751, 778, 779, 780, 781,
			782, 783, 784, 785, 812, 813, 814, 815,
			816, 817, 818, 819, 846, 847, 848, 849,
			850, 851, 852, 853, 880, 881, 882, 883,
			884, 885, 886, 887, 914, 915, 916, 917,
			918, 919, 920, 921, 948, 949, 950, 951,
			952, 953, 954, 955, 982, 983, 984, 985,
			986, 987, 988, 989, 990, 991, 992, 993, 994, 995, 996, 997, 998, 999, 1000, 1001, 1002, 1003, 1004, 1005, 1006, 1007, 1008, 1009, 1010, 1011, 1012, 1013, 1014, 1015, 1016, 1017, 1018, 1019,
			1020, 1021, 1022, 1023, 1024, 1025, 1026, 1027, 1028, 1029, 1030, 1031, 1032, 1033, 1034, 1035, 1036, 1037, 1038, 1039, 1040, 1041, 1042, 1043, 1044, 1045, 1046, 1047, 1048, 1049, 1050, 1051, 1052, 1053,
			1054, 1055, 1056, 1057, 1058, 1059, 1060, 1061, 1062, 1063, 1064, 1065, 1066, 1067, 1068, 1069, 1070, 1071, 1072, 1073, 1074, 1075, 1076, 1077, 1078, 1079, 1080, 1081, 1082, 1083, 1084, 1085, 1086, 1087,
		},
	},
}

// GameMapCreate create a new game map
func GameMapCreate(ctx context.Context, conn *pgxpool.Pool, gameMap *GameMap) error {
	q := `
		INSERT INTO 
			game_maps (
				name, 
				image_url, 
				width, 
				height, 
				cells_x, 
				cells_y, 
				top_pixels, 
				left_pixels, 
				scale, 
				disabled_cells,
				max_spawns
			)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING 
			id,
			name, 
			image_url, 
			width, 
			height, 
			cells_x, 
			cells_y, 
			top_pixels, 
			left_pixels, 
			scale,
			disabled_cells,
			max_spawns
		
	`
	err := pgxscan.Get(ctx, conn, gameMap, q,
		gameMap.Name,
		gameMap.ImageUrl,
		gameMap.Width,
		gameMap.Height,
		gameMap.CellsX,
		gameMap.CellsY,
		gameMap.TopPixels,
		gameMap.LeftPixels,
		gameMap.Scale,
		gameMap.DisabledCells,
		gameMap.MaxSpawns,
	)
	if err != nil {
		return terror.Error(err)
	}

	return nil
}
