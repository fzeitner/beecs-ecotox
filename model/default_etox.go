package model

import (
	"github.com/mlange-42/ark-tools/app"
	"github.com/mlange-42/ark/ecs"
	"github.com/mlange-42/beecs/globals"
	"github.com/mlange-42/beecs/params"
	"github.com/mlange-42/beecs/sys"
)

// DefaultEtox sets up the default beecs_ecotox model with the standard sub-models.
//
// If the argument m is nil, a new model instance is created.
// If it is non-nil, the model is reset and re-used, saving some time for initialization and memory allocation.
func DefaultEtox(p params.Params, pe params.ParamsEtox, app *app.App) *app.App {

	// Add parameters and other resources

	app = initializeModelEtox(p, pe, app)

	// Initialization
	app.AddSystem(&sys.InitStore{})          // unchanged to beecs
	app.AddSystem(&sys.InitCohorts{})        // unchanged to beecs
	app.AddSystem(&sys.InitPopulation{})     // unchanged to beecs
	app.AddSystem(&sys.InitPatchesList{})    // unchanged to beecs
	app.AddSystem(&sys.InitForagingPeriod{}) // unchanged to beecs
	app.AddSystem(&sys.InitEtox{})           // inits all the changes necessary for the beecs_ecotox submodels

	// Sub-models
	app.AddSystem(&sys.CalcAff{})            // unchanged to beecs
	app.AddSystem(&sys.CalcForagingPeriod{}) // unchanged to beecs
	app.AddSystem(&sys.ReplenishPatches{})   // unchanged to beecs
	app.AddSystem(&sys.PPPApplication{})     // introduced calculation of PPP exposure at patches analogous to BEEHAVE_ecotox

	app.AddSystem(&sys.MortalityCohorts{})     // unchanged to beecs
	app.AddSystem(&sys.MortalityCohortsEtox{}) // introduced ETOXMortality as an additional process for all cohorts
	app.AddSystem(&sys.AgeCohorts{})           // unchanged to beecs
	app.AddSystem(&sys.EggLaying{})            // unchanged to beecs
	app.AddSystem(&sys.TransitionForagers{})   // unchanged to beecs

	app.AddSystem(&sys.CountPopulation{}) // added here to reflect position in original model, necessary to capture mortality effects of cohorts on broodcare and foraging
	app.AddSystem(&sys.BroodCare{})       // Moved after the first countingproc to resemble the original model further, as counting twice is inevitable because of ETOXmortality processes.

	app.AddSystem(&sys.NewCohorts{})      // unchanged to beecs
	app.AddSystem(&sys.CountPopulation{}) // added here to reflect position in original model (miteproc), necessary to capture new Cohorts for foraging

	app.AddSystem(&sys.ForagingEtox{})          // introduced the uptake of PPP into foragers and the hive through contaminated honey/pollen; would be far too tedious to decouple this from the normal foraging submodel
	app.AddSystem(&sys.MortalityForagers{})     // unchanged to beecs
	app.AddSystem(&sys.MortalityForagersEtox{}) // introduced ETOXMortality as an additional process for foragers after normal foraging mortality, analogous to BEEHAVE_ecotox

	app.AddSystem(&sys.CountPopulation{})   // necessary here because of food comsumption in the next steps
	app.AddSystem(&sys.PollenConsumption{}) // unchanged to beecs
	app.AddSystem(&sys.HoneyConsumption{})  // unchanged to beecs
	app.AddSystem(&sys.EtoxStorages{})      // regulates in-hive exposure and fate of PPP and the newly introduced honey compartiments

	app.AddSystem(&sys.FixedTermination{})

	return app
}

// WithSystems sets up a beecs model with the given systems instead of the default ones.
//
// If the argument m is nil, a new model instance is created.
// If it is non-nil, the model is reset and re-used, saving some time for initialization and memory allocation.
func WithSystemsEtox(p params.Params, pe params.ParamsEtox, sys []app.System, app *app.App) *app.App {

	app = initializeModelEtox(p, pe, app)

	for _, s := range sys {
		app.AddSystem(s)
	}

	return app
}

func initializeModelEtox(p params.Params, pe params.ParamsEtox, a *app.App) *app.App {
	if a == nil {
		a = app.New()
	} else {
		a.Reset()
	}

	p.Apply(&a.World)
	pe.Apply(&a.World)

	factory := globals.NewForagerFactory(&a.World)
	ecs.AddResource(&a.World, &factory)

	stats := globals.PopulationStats{}
	ecs.AddResource(&a.World, &stats)

	consumptionStats := globals.ConsumptionStats{}
	ecs.AddResource(&a.World, &consumptionStats)

	foragingStats := globals.ForagingStats{}
	ecs.AddResource(&a.World, &foragingStats)

	return a
}
