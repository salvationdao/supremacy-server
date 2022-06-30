package main

import (
	"encoding/csv"
	"fmt"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"os"
	"server/db/boiler"
	"server/devtool/types"
	"strings"
)

func (dt DevTool) SyncMechs() error {
	f, err := os.OpenFile("./temp-sync/supremacy-static-data/mech_models.csv", os.O_RDONLY, 0755)
	if err != nil {
		return err
	}

	defer f.Close()

	r := csv.NewReader(f)

	if _, err := r.Read(); err != nil {
		return err
	}

	records, err := r.ReadAll()
	if err != nil {
		return err
	}

	_, err = dt.db.Exec(
		`
			ALTER TABLE mech_models DROP CONSTRAINT mech_model_default_chassis_skin_id_fkey;
			ALTER TABLE mech_models ADD CONSTRAINT mech_model_default_chassis_skin_id_fkey FOREIGN KEY (default_chassis_skin_id) REFERENCES blueprint_mech_skin(id) ON UPDATE CASCADE;

			ALTER TABLE blueprint_mech_skin DROP CONSTRAINT blueprint_chassis_skin_mech_model_fkey;
			ALTER TABLE blueprint_mech_skin ADD CONSTRAINT blueprint_chassis_skin_mech_model_fkey FOREIGN KEY (mech_model) REFERENCES mech_models(id) ON UPDATE CASCADE;

			ALTER TABLE mech_skin DROP CONSTRAINT chassis_skin_blueprint_id_fkey;
			ALTER TABLE mech_skin ADD CONSTRAINT chassis_skin_blueprint_id_fkey FOREIGN KEY (blueprint_id) REFERENCES blueprint_mech_skin(id) ON UPDATE CASCADE;

			ALTER TABLE mech_skin DROP CONSTRAINT chassis_skin_mech_model_fkey;
			ALTER TABLE mech_skin ADD CONSTRAINT chassis_skin_mech_model_fkey FOREIGN KEY (mech_model) REFERENCES mech_models(id) ON UPDATE CASCADE;

			ALTER TABLE mechs DROP CONSTRAINT chassis_model_id_fkey;
			ALTER TABLE mechs ADD CONSTRAINT chassis_model_id_fkey FOREIGN KEY (model_id) REFERENCES mech_models(id) ON UPDATE CASCADE;

			ALTER TABLE blueprint_mechs DROP CONSTRAINT blueprint_chassis_model_id_fkey;
			ALTER TABLE blueprint_mechs ADD CONSTRAINT blueprint_chassis_model_id_fkey FOREIGN KEY (model_id) REFERENCES mech_models(id) ON UPDATE CASCADE;

			ALTER TABLE blueprint_mechs DROP CONSTRAINT blueprint_chassis_brand_id_fkey;
			ALTER TABLE blueprint_mechs ADD CONSTRAINT blueprint_chassis_brand_id_fkey FOREIGN KEY (brand_id) REFERENCES brands(id) ON UPDATE CASCADE;
			`,
	)

	var MechModels []types.MechModel
	for _, record := range records {
		mechModel := &types.MechModel{
			ID:                   record[0],
			Label:                record[1],
			DefaultChassisSkinID: record[3],
			BrandID:              null.StringFrom(record[4]),
			MechType:             record[5],
		}

		MechModels = append(MechModels, *mechModel)
	}

	for _, mechModel := range MechModels {

		model, err := boiler.MechModels(boiler.MechModelWhere.Label.EQ(mechModel.Label)).One(dt.db)
		if err != nil {
			fmt.Println("ERROR")
			continue
		}

		if strings.EqualFold(mechModel.ID, model.ID) {
			continue
		}

		model.ID = mechModel.ID

		_, err = model.Update(dt.db, boil.Infer())
		if err != nil {
			fmt.Println("ERROR")
			continue
		}
	}

	fmt.Println("Finish syncing mech models")

	return nil
}
