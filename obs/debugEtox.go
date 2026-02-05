package obs

import (
	"github.com/mlange-42/ark/ecs"
	"github.com/mlange-42/beecs/globals"
)

// Debug is a row observer for several colony structure variables,
// using the same names as the original BEEHAVE_ecotox implementation.
//
// Primarily meant for validation of beecs against BEEHAVE_ecotox.
type DebugEtox struct {
	pop      *globals.PopulationStats
	popEtox  *globals.PopulationStatsEtox
	stores   *globals.Stores
	foraging *globals.ForagingPeriod
	data     []float64
}

func (o *DebugEtox) Initialize(w *ecs.World) {
	o.pop = ecs.GetResource[globals.PopulationStats](w)
	o.popEtox = ecs.GetResource[globals.PopulationStatsEtox](w)

	o.stores = ecs.GetResource[globals.Stores](w)
	o.foraging = ecs.GetResource[globals.ForagingPeriod](w)
	o.data = make([]float64, len(o.Header()))
}
func (o *DebugEtox) Update(w *ecs.World) {}
func (o *DebugEtox) Header() []string {
	return []string{"DailyForagingPeriod", "HoneyEnergyStore", "PollenStore_g", "TotalEggs", "TotalLarvae", "TotalPupae", "TotalIHBees", "TotalForagers", "ETOX_Mean_Dose_Larvae", "ETOX_Mean_Dose_IHBee", "ETOX_Mean_Dose_Forager", "ETOX_Cum_Dose_Larvae", "ETOX_Cum_Dose_IHBee", "ETOX_Cum_Dose_Forager", "TotalPop"}
}
func (o *DebugEtox) Values(w *ecs.World) []float64 {
	o.data[0] = float64(o.foraging.SecondsToday)
	o.data[1] = o.stores.Honey
	o.data[2] = o.stores.Pollen

	o.data[3] = float64(o.pop.WorkerEggs)
	o.data[4] = float64(o.pop.WorkerLarvae)
	o.data[5] = float64(o.pop.WorkerPupae)
	o.data[6] = float64(o.pop.WorkersInHive)
	o.data[7] = float64(o.pop.WorkersForagers)

	o.data[8] = float64(o.popEtox.MeanDoseLarvae)
	o.data[9] = float64(o.popEtox.MeanDoseIHBees)
	o.data[10] = float64(o.popEtox.MeanDoseForager)

	o.data[11] = float64(o.popEtox.CumDoseLarvae)
	o.data[12] = float64(o.popEtox.CumDoseIHBees)
	o.data[13] = float64(o.popEtox.CumDoseForagers)

	o.data[14] = float64(o.pop.WorkerEggs + o.pop.WorkerLarvae + o.pop.WorkerPupae + o.pop.WorkersInHive + o.pop.WorkersForagers + o.pop.DroneEggs + o.pop.DroneLarvae + o.pop.DronePupae + o.pop.DronesInHive)

	return o.data
}
