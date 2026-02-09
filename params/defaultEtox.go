package params

import (
	"bytes"
	"encoding/json"
	"os"

	"github.com/mlange-42/ark/ecs"
)

// ParamsEtox is an interface for the beecs_ecotox parameter sets.
type ParamsEtox interface {
	// Apply the parameters to a world.
	Apply(world *ecs.World)
	// FromJSON fills the parameter set with values from a JSON file.
	FromJSONFile(path string) error
	// FromJSON fills the parameter set with values from a JSON file.
	FromJSON(data []byte) error
}

// DefaultParamsEtox contains all default parameters of BEEHAVE_ecotox.
//
// DefaultParamsEtox implements [ParamsEtox].
type DefaultParamsEtox struct {
	PPPApplication PPPApplication
	PPPToxicity    PPPToxicity
}

// DefaultEtox returns the complete default parameter set for beecs_ecotox. ReworkedThermoEtox, RealisticStoch and the two fixes are additions created by me.
func DefaultEtox() DefaultParamsEtox {
	return DefaultParamsEtox{
		PPPApplication: PPPApplication{
			Application:               false, // Determines whether there is an application at all (and turns on/off the necessary code).
			ForagerImmediateMortality: false, // Determines whether it is taken into account that foragers can die from exposure during a foraging trip which would reduce the amount of compound brought back to the hive.
			DegradationHoney:          false, // Determines whether the compound in the honey (within the hive) does degrade or not.
			ContactSum:                false, // Determines whether contact exposure should be summed up per visit to a patch (true) or if the mean should be calculated whenever a new patch is visited (false).
			ContactExposureOneDay:     false, // Determines whether contact exposure should only be possible on the day of application.
			FixedNectarPollenRatio:    false, // Determines if the PPP concentration in pollen should be estimated based on PPP conc in nectar and a constant factor.

			RealisticStoch:     false, // Determines whether stochstic death for low numbers of IHbees in one cohort shall be made more realistic by calculating a chance for each bee.
			ReworkedThermoETOX: false, // Determines whether thermoregulation energy shall be taken in equally by all adult bees (True, new version) or if one cohort/squad shall take it all (false; Netlogo version).
			Nursebeefix:        true,  // Added by F. Zeitner, not implemented in NetLogo. Determines whether the nurse bee intake from BEEHAVE_ecotox's nursebeefactors shall be added to IHbees instead of dissipating.
			HSUfix:             true,  // Added by F. Zeitner, not implemented in NetLogo. Determines if the PPP lost to the second call of HSuptake when unloading nectar shall be redirected to IHbees (true) instead of dissipating.

			PPPname:                "No applications", // Identifier for the PPP used.
			PPPconcentrationNectar: 990,               // PPP concentration in nectar [mug/kg].
			PPPconcentrationPollen: 26631,             // PPP concentration in pollen [mug/kg].
			PPPcontactExposure:     0.3,               // PPP concentration for contact exposure on patch [kg/ha].

			AppDay:           189,   // Day of the year in which application starts [d].
			ExposurePeriod:   8,     // Duration of exposure happening (irrespective of DT50) [d].
			SpinupPhase:      0,     // Number of years before exposure starts (to stabilize colony; 0 = first year) [y].
			ExposurePhase:    3,     // Number of years in which exposure takes place [y].
			DT50:             1000., // Whole plant DT50 from residue studies [d].
			RUD:              21.,   // Residue per Unit Dose  [(ha*mg)/(kg*kg)].
			DT50honey:        60.,   // Honey DT50 [d].
			EtoxNecPolFactor: 26.9,  // // Constant factor to estimate PPP in pollen based on nectar.

			ETOXDensityOfHoney: 1.4, // The density of honey is 1.4 [kg/l].
		},
		PPPToxicity: PPPToxicity{
			// most of these default values are identical to the fenoxycarb parameters estimated in the suppl. material
			// of the Preuss et al. (2022); just the ForagerContactLD50 of 0.6 is nowhere to be found. It is, however, the
			// default value of BEEHAVE_ecotox, therefore it was kept as the default value for this model, too.
			// it is unclear of these parameters in this combination actually belong to any one substance overall.
			ForagerOralLD50:  1000., // Lethal oral dose for 50% mortality of foragers [µg/bee].
			ForagerOralSlope: 100.,  // Slope of the dose-response relationship (forager, oral) [ ].
			HSuptake:         0.1,   // Uptake of a given percentage of ai in the honey stomach by the forager bees

			ForagerContactLD50:  0.6,  // Lethal dose for 50% of foragers via contact exposure [µg/bee]
			ForagerContactSlope: 1.08, // Slope of the dose-response relationship (forager, contact) [ ]

			LarvaeOralLD50:  0.0014, // Lethal oral dose for 50% mortality of larvae [µg/larvae]
			LarvaeOralSlope: 1.6,    // Slope of the dose-response relationship (larvae, oral) [ ]; A log-normal dose-response curve is implemented

			NursebeesNectar: 0.25, // Factor describing the filter effect of nurse bees for nectar [ ]
			NursebeesPollen: 1.,   // Factor describing the filter effect of nurse bees for pollen [ ]
		},
	}
}

// FromJSONFile fills the parameter set with values from a JSON file.
//
// Only values present in the file are overwritten,
// all other values remain unchanged.
func (p *DefaultParamsEtox) FromJSONFile(path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	return p.FromJSON(content)
}

// FromJSON fills the parameter set with values from JSON.
//
// Only values present in the file are overwritten,
// all other values remain unchanged.
func (p *DefaultParamsEtox) FromJSON(data []byte) error {
	reader := bytes.NewReader(data)
	decoder := json.NewDecoder(reader)
	decoder.DisallowUnknownFields()
	return decoder.Decode(p)
}

// Apply the parameters to a world by adding them as resources.
func (p *DefaultParamsEtox) Apply(world *ecs.World) {
	pCopy := *p

	// Resources
	ecs.AddResource(world, &pCopy.PPPApplication)
	ecs.AddResource(world, &pCopy.PPPToxicity)
}
