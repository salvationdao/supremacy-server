package supermigrate

import (
	"fmt"
	"math/rand"

	"github.com/gosimple/slug"
)

func DefaultTemplateLabelSlug(name string) (string, string) {
	switch name {
	case "Alex":

	}
	return "", ""
}

func TemplateLabelSlug(brand, model, submodel string) (string, string) {
	label := fmt.Sprintf("%s %s %s", brand, model, submodel)
	return label, label
}

func BlueprintMechLabelSlug(brand, model, submodel, name string) (string, string) {
	label := fmt.Sprintf("%s %s %s %s", brand, model, submodel, name)
	return label, slug.Make(label)
}

func BlueprintWeaponLabelSlug(brand, weaponName string) (string, string) {
	label := fmt.Sprintf("%s %s", brand, weaponName)
	return label, slug.Make(label)
}

func BlueprintModuleLabelSlug(brand, moduleName string) (string, string) {
	label := fmt.Sprintf("%s %s", brand, moduleName)
	return label, slug.Make(label)
}

func BlueprintChassisLabelSlug(brand, model, submodel string) (string, string) {
	label := fmt.Sprintf("%s %s %s Chassis", brand, model, submodel)
	return label, slug.Make(label)
}

func MechLabelSlug(brand, model, submodel, name string) (string, string) {
	label := fmt.Sprintf("%s %s %s %s", brand, model, submodel, name)
	return label, slug.Make(fmt.Sprintf("%s#%d", label, 1000+rand.Intn(8999)))
}

func WeaponLabelSlug(brand, weaponName string) (string, string) {
	label := fmt.Sprintf("%s %s", brand, weaponName)
	return label, slug.Make(fmt.Sprintf("%s#%d", label, 1000+rand.Intn(8999)))
}

func ModuleLabelSlug(brand, moduleName string) (string, string) {
	label := fmt.Sprintf("%s %s", brand, moduleName)
	return label, slug.Make(fmt.Sprintf("%s#%d", label, 1000+rand.Intn(8999)))
}

func ChassisLabelSlug(brand, model, submodel string) (string, string) {
	label := fmt.Sprintf("%s %s %s Chassis", brand, model, submodel)
	return label, slug.Make(fmt.Sprintf("%s#%d", label, 1000+rand.Intn(8999)))
}
