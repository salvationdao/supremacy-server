package db

import (
	"server"

	"github.com/georgysavva/scany/pgxscan"
	"github.com/ninja-software/terror/v2"
	"golang.org/x/net/context"
)

// CreateStream created a new stream
func CreateStream(ctx context.Context, conn Conn, stream *server.Stream) error {
	q := `
		INSERT INTO
			stream_list (id, name, url, region, resolution, bit_rates_k_bits, user_max, users_now, active, status, latitude, longitude)
		VALUES
			($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING
		id, name, url, region, resolution, bit_rates_k_bits, user_max, users_now, active, status
	`

	err := pgxscan.Get(ctx, conn, stream, q, stream.ID, stream.Name, stream.Url, stream.Region, stream.Resolution, stream.BitRatesKBits, stream.UserMax, stream.UsersNow, stream.Active, stream.Status, stream.Latitude, stream.Longitude)
	if err != nil {
		return terror.Error(err)
	}

	return nil
}

func GetStreamList(ctx context.Context, conn Conn) ([]*server.Stream, error) {
	var streamList []*server.Stream
	q := `SELECT * FROM stream_list`

	err := pgxscan.Select(ctx, conn, &streamList, q)
	if err != nil {
		return nil, terror.Error(err)
	}

	return streamList, nil
}
