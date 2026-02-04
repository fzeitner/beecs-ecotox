package sys

import (
	"math"
	"math/rand/v2"

	"github.com/mlange-42/ark-tools/resource"
	"github.com/mlange-42/ark/ecs"
	"github.com/mlange-42/beecs/globals"
	"github.com/mlange-42/beecs/params"
	"github.com/mlange-42/beecs/util"
)

// MortalityCohortsEtox applies ETOX-related mortality to all in-hive cohorts
// analogously to BEEHAVE_ecotox. This is part of beecs_ecotox/nursebeecs_ecotox only.
type MortalityCohortsEtox struct {
	workerMort *params.WorkerMortality
	droneMort  *params.DroneMortality

	larvae     *globals.Larvae
	larvaeEtox *globals.LarvaeEtox
	inHive     *globals.InHive
	inHiveEtox *globals.InHiveEtox
	popStats   *globals.PopulationStatsEtox

	etox  *params.PPPApplication
	toxic *params.PPPToxicity

	rng *resource.Rand
}

func (s *MortalityCohortsEtox) Initialize(w *ecs.World) {
	s.workerMort = ecs.GetResource[params.WorkerMortality](w)
	s.droneMort = ecs.GetResource[params.DroneMortality](w)

	s.larvae = ecs.GetResource[globals.Larvae](w)
	s.larvaeEtox = ecs.GetResource[globals.LarvaeEtox](w)
	s.inHive = ecs.GetResource[globals.InHive](w)
	s.inHiveEtox = ecs.GetResource[globals.InHiveEtox](w)
	s.popStats = ecs.GetResource[globals.PopulationStatsEtox](w)

	s.etox = ecs.GetResource[params.PPPApplication](w)
	s.toxic = ecs.GetResource[params.PPPToxicity](w)

	s.rng = ecs.GetResource[resource.Rand](w)
}

func (s *MortalityCohortsEtox) Update(w *ecs.World) {
	s.applyMortalityEtox(s.larvae.Workers, s.larvaeEtox.WorkerCohortDose, s.toxic.LarvaeOralSlope, s.toxic.LarvaeOralLD50)
	s.applyMortalityEtox(s.larvae.Drones, s.larvaeEtox.DroneCohortDose, s.toxic.LarvaeOralSlope, s.toxic.LarvaeOralLD50)

	s.applyMortalityEtox(s.inHive.Workers, s.inHiveEtox.WorkerCohortDose, s.toxic.ForagerOralSlope, s.toxic.ForagerOralLD50)

	s.applyMortalityEtox(s.inHive.Drones, s.inHiveEtox.DroneCohortDose, s.toxic.ForagerOralSlope, s.toxic.ForagerOralLD50)

	s.popStats.Reset() // resets cumulative and mean doses for the timestep
}

func (s *MortalityCohortsEtox) Finalize(w *ecs.World) {}

func (s *MortalityCohortsEtox) applyMortalityEtox(coh []int, dose []float64, slope float64, LD50 float64) {
	r := rand.New(s.rng)
	for i := range coh {
		num := coh[i]
		toDie := 0
		if dose[i] > 1e-20 { // simple dose response relationship for all larvae/IHBees/drones
			num = coh[i]
			ldx := (1 - (1 / (1 + math.Pow((dose[i]/LD50), slope))))
			if ldx > 0.99 { // introduced this because NetLogo-version behaves the same way. This makes it much less likely to have single digit cohorts left over after very lethal PPP events
				ldx = 1
			}
			if s.etox.RealisticStoch { // this is deactivated by default and not part of BEEHAVE_ecotox, but found to make sense in some cases
				if num > 100 { // introduced this to improve stochasticity for cohorts with low numbers of individuals
					toDie = int((float64(num) * ldx))
				} else {
					i := 0
					for i < num {
						if r.Float64() < ldx {
							toDie++
						}
						i++
					}
				}
			} else {
				toDie = int((float64(num) * ldx))
			}
		}
		coh[i] = util.MaxInt(0, num-toDie)
		dose[i] = 0. // doses get reset to 0 after the mortality check in every timestep, only dose from previous day is ever relevant
	}
}
