package db

import (
	"server"

	"github.com/georgysavva/scany/pgxscan"
	"github.com/ninja-software/terror/v2"
	"golang.org/x/net/context"
)

// WarMachineCreate create a list of war machine
func WarMachineCreate(ctx context.Context, conn Conn, warMachine *server.WarMachine) error {
	q := `
		INSERT INTO 
			war_machines (id, name, base_health_point, base_shield_point)
		VALUES
			($1, $2, $3, $4)
		RETURNING
			id, name, base_health_point, base_shield_point;
	`

	err := pgxscan.Get(ctx, conn, warMachine, q,
		warMachine.ID,
		warMachine.Name,
		warMachine.BaseHealthPoint,
		warMachine.BaseShieldPoint,
	)
	if err != nil {
		return terror.Error(err)
	}
	return nil
}

func WarMachineAll(ctx context.Context, conn Conn) ([]*server.WarMachine, error) {
	result := []*server.WarMachine{}

	q := `
		SELECT * FROM war_machines
	`

	err := pgxscan.Select(ctx, conn, &result, q)
	if err != nil {
		return nil, terror.Error(err)
	}

	return result, nil
}
