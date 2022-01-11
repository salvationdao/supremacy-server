package passport

import (
	"server"

	"github.com/gofrs/uuid"
)

func (pp *Passport) GetWarMachines() []*server.WarMachine {
	return fakeWarMachines
}

var fakeWarMachines = []*server.WarMachine{
	{
		ID:        server.WarMachineID(uuid.Must(uuid.FromString("b0d0cb39-bd84-478b-aea9-29f7482260af"))),
		Name:      "Zeus",
		FactionID: FakeFactions[0].ID,
		WarMachinePosition: &server.WarMachinePosition{
			Rotation: 270,
			Position: &server.Vector3{
				X: 70,
				Y: 70,
				Z: 0,
			},
		},
	},
	{
		ID:        server.WarMachineID(uuid.Must(uuid.FromString("9ba60e86-fc2f-4001-b3b4-914173b82ac4"))),
		Name:      "Poseidon",
		FactionID: FakeFactions[1].ID,
		WarMachinePosition: &server.WarMachinePosition{
			Rotation: 220,
			Position: &server.Vector3{
				X: 90,
				Y: 86,
				Z: 0,
			},
		},
	},
	{
		ID:        server.WarMachineID(uuid.Must(uuid.FromString("5aa9e652-ea2d-4bab-ad81-855e6c5c3bab"))),
		Name:      "Hera",
		FactionID: FakeFactions[2].ID,
		WarMachinePosition: &server.WarMachinePosition{
			Rotation: 6,
			Position: &server.Vector3{
				X: 50,
				Y: 90,
				Z: 0,
			},
		},
	},
	{
		ID:        server.WarMachineID(uuid.Must(uuid.FromString("970cb67f-305b-473f-a8c1-12bb78125bb0"))),
		Name:      "Athena",
		FactionID: FakeFactions[0].ID,
		WarMachinePosition: &server.WarMachinePosition{
			Rotation: 110,
			Position: &server.Vector3{
				X: 105,
				Y: 98,
				Z: 0,
			},
		},
	},
	{
		ID:        server.WarMachineID(uuid.Must(uuid.FromString("31e19f3d-2670-4e6c-b635-7e80940ef2f9"))),
		Name:      "Hercules",
		FactionID: FakeFactions[1].ID,
		WarMachinePosition: &server.WarMachinePosition{
			Rotation: 320,
			Position: &server.Vector3{
				X: 76,
				Y: 120,
				Z: 0,
			},
		},
	},
	{
		ID:        server.WarMachineID(uuid.Must(uuid.FromString("757f52c4-bb78-4e78-8240-66bc05cbf37b"))),
		Name:      "Hephaestus",
		FactionID: FakeFactions[2].ID,
		WarMachinePosition: &server.WarMachinePosition{
			Rotation: 92,
			Position: &server.Vector3{
				X: 60,
				Y: 100,
				Z: 0,
			},
		},
	},
}
