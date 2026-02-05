package params

// parameters for the application of pesticides.
type PPPApplication struct {
	Application               bool // Determines if there is an application at all at any point in the model and if the _ecotox-module should be turned on for all purposes
	ForagerImmediateMortality bool // Determines whether it is taken into account that foragers can die from exposure during a foraging trip which would reduce the amount of compound brought back to the hive.
	DegradationHoney          bool // Determines whether the compound in the honey (within the hive) does degrade or not. This does impact the in-hive toxicity of the compound,
	ContactSum                bool // Determines whether contact exposures of different flower visits shall be summed up.
	ContactExposureOneDay     bool // Determines whether contact exposure shall only be relevant on the one day of application
	FixedNectarPollenRatio    bool // Determines if the PPP concentration in pollen should be estimated based on PPP conc in nectar and a constant factor.

	RealisticStoch     bool // Determines whether stochstic death for low numbers of IHbees in one cohort shall be made more realistic by calculating a chance for each bee
	ReworkedThermoETOX bool // Determines whether thermoregulation energy shall be taken in equally by all adult bees (True, new version) or if one cohort/squad shall take it all (false; Netlogo version)
	Nursebeefix        bool // Determines if the nurse bee intake from BEEHAVE_ecotox's nursebeefactors shall be added to IHbees instead of dissipating
	HSUfix             bool // Determines if the PPP lost to the second call of HSuptake when unloading nectar shall be redirected to IHbees (true) insted of dissipating

	PPPname                string  // Identifier for the PPP used.
	PPPconcentrationNectar float64 // PPP concentration in nectar [mug/kg].
	PPPconcentrationPollen float64 // PPP concentration in pollen [mug/kg].
	PPPcontactExposure     float64 // PPP concentration for contact exposure on patch [kg/ha].

	AppDay           int     // Day of the year in which application starts [d].
	ExposurePeriod   int     // Duration of exposure happening (irrespective of DT50) [d].
	SpinupPhase      int     // Number of years before exposure starts (to stabilize colony; 0 = first year) [y].
	ExposurePhase    int     // Number of years in which exposure takes place [y].
	DT50             float64 // Whole plant DT50 from residue studies [d].
	RUD              float64 // Residue per Unit Dose  [(ha*mg)/(kg*kg)].
	DT50honey        float64 // Honey DT50 [d].
	EtoxNecPolFactor float64 // Constant factor to estimate PPP in pollen based on nectar.

	ETOXDensityOfHoney float64 // The density of honey is 1.4 [kg/l].
}

// parameters for uptake and toxicity of the applied pesticide to foragers and cohorts.
type PPPToxicity struct {
	ForagerOralLD50  float64 // Lethal oral dose for 50% mortality of foragers [µg/bee].
	ForagerOralSlope float64 // Slope of the dose-response relationship (forager, oral) [ ].
	HSuptake         float64 // Uptake of a given percentage of ai in the honey stomach by the forager bees.

	ForagerContactLD50  float64 // Lethal dose for 50% of foragers via contact exposure [µg/bee].
	ForagerContactSlope float64 // Slope of the dose-response relationship (forager, contact) [ ].

	LarvaeOralLD50  float64 // Lethal oral dose for 50% mortality of larvae [µg/larvae].
	LarvaeOralSlope float64 // Slope of the dose-response relationship (larvae, oral) [ ]; A log-normal dose-response curve is implemented.

	NursebeesNectar float64 // Factor describing the filter effect of nurse bees for nectar [ ].
	NursebeesPollen float64 // Factor describing the filter effect of nurse bees for pollen [ ].
}
