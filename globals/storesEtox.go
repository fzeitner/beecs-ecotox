package globals

// globals necessary for the _ecotox additions of water foraging, ecotoxicological variables and honey storage compartimentation
// none of this is used with the default ecotox settings and should not, because of a lack of testing atm

// StoragesEtox tracks the current concentration of PPP inside of the compartimentarized honey stores added in BEEHAVE_ecotox.
// It also traks some PPP-related ovserving variables as well as the energy necessary for thermoregulation.
type StoragesEtox struct {
	PPPInHivePollenConc float64 // Concentration of PPP currently in stored pollen [mug/g].
	EtoxEnergyThermo    float64 // Energy needed for Thermoregulation of hive/brood made global for Etox_consumption purposes.

	EtoxHoneyEnergy        []float64 // Energy in the honey cells [kJ]. Slice from todays uncapped honey cells [0] to capped honey cells [5].
	EtoxHoneyConcentration []float64 // Average concentration of pesticide in the honey cells [µg/kJ]. Slice from todays uncapped honey cells [0] to capped honey cells [5].

	Pollenconcbeforeeating float64 // added for bugfixing
	Nectarconcbeforeeating float64 // added for bugfixing

	PPPpollenTotal float64 // total amount of PPP in pollen stores this timestep
	PPPhoneyTotal  float64 // total amount of PPP in honey stores this timestep
	PPPTotal       float64 // total amount of PPP in all stores this timestep
}

// PPPFate tracks the total amount of PPP that flows into the respective PPP-sinks and was used to create
// PPP mass balances (see examples/beecs_ecotox)
type PPPFate struct {
	TotalPPPforaged      float64 // Total amount of PPP foraged by all foragers; used to create a mass balance
	PPPhoneyStores       float64 // Total amount of PPP that ends up in honey stores after being foraged
	PPPpollenStores      float64 // Total amount of PPP that ends up in pollen stores after being foraged
	PPPforagersImmediate float64 // amount of PPP taken in by foragers immediately while foraging
	ForagerDiedInFlight  float64 // amount of PPP "lost" by foragers dying midflight of bringing back a PPPload

	PPPforagersinHive float64 // amount of PPP taken in by foragers via normal in-hive oral uptake
	PPPforagersTotal  float64 // total amount of PPP taken in by foragers via all oral routes
	PPPIHbees         float64 // amount of PPP taken in by IHbees
	PPPNurses         float64 // amount of PPP taken in by nurse bees
	PPPlarvae         float64 // amount of PPP taken in by larvae
	PPPdrones         float64 // amount of PPP taken in by drones
	PPPdlarvae        float64 // amount of PPP taken in by drone larvae
}
