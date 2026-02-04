package sys

import (
	"fmt"
	"math"
	"math/rand/v2"

	"github.com/mlange-42/ark-tools/resource"
	"github.com/mlange-42/ark/ecs"
	"github.com/mlange-42/beecs/comp"
	"github.com/mlange-42/beecs/enum/activity"
	"github.com/mlange-42/beecs/globals"
	"github.com/mlange-42/beecs/params"
	"github.com/mlange-42/beecs/util"
	"gonum.org/v1/gonum/stat/distuv"
)

// Foraging performs the complete foraging process of each day.
// It potentially performs multiple foraging rounds per day.
type ForagingEtox struct {
	rng  *rand.Rand
	time *resource.Tick

	foragerParams      *params.Foragers
	forageParams       *params.Foraging
	handlingTimeParams *params.HandlingTime
	danceParams        *params.Dance
	energyParams       *params.EnergyContent
	storeParams        *params.Stores

	etox          *params.PPPApplication
	toxic         *params.PPPToxicity
	nursingParams *params.Nursing

	foragingStats *globals.ForagingStats
	foragePeriod  *globals.ForagingPeriod
	stores        *globals.Stores
	storesEtox    *globals.StoragesEtox
	pppfate       *globals.PPPFate
	pop           *globals.PopulationStats
	newCohorts    *globals.NewCohorts
	aff           *globals.AgeFirstForaging
	factory       *globals.ForagerFactory

	patches        []patchCandidateEtox
	toRemove       []ecs.Entity
	resting        []ecs.Entity
	dances         []ecs.Entity
	searches       []ecs.Entity
	recruits       []ecs.Entity
	toAdd          []ecs.Entity
	foragerShuffle []ecs.Entity

	ageMapper             *ecs.Map1[comp.Age]
	patchResourceMapper   *ecs.Map1[comp.Resource]
	patchVisitsMapper     *ecs.Map2[comp.Resource, comp.Visits]
	patchDanceMapper      *ecs.Map2[comp.Resource, comp.Dance]
	patchTripMapper       *ecs.Map1[comp.Trip]
	patchMortalityMapper  *ecs.Map1[comp.Mortality]
	patchConfigMapper     *ecs.Map2[comp.PatchProperties, comp.Trip]
	patchConfigMapperEtox *ecs.Map3[comp.PatchProperties, comp.PatchPropertiesEtox, comp.Trip]
	foragerMapper         *ecs.Map2[comp.Activity, comp.KnownPatch]
	foragerLoadPPPMapper  *ecs.Map6[comp.Activity, comp.KnownPatch, comp.Milage, comp.NectarLoad, comp.EtoxLoad, comp.PPPExpo]
	pppExpoAdder          *ecs.Map2[comp.PPPExpo, comp.EtoxLoad]

	activityFilter       *ecs.Filter1[comp.Activity]
	ageFilter            *ecs.Filter1[comp.Age]
	loadFilter           *ecs.Filter3[comp.Activity, comp.NectarLoad, comp.EtoxLoad]
	loadexpoFilter       *ecs.Filter4[comp.Activity, comp.NectarLoad, comp.EtoxLoad, comp.PPPExpo]
	foragerFilter        *ecs.Filter3[comp.Activity, comp.KnownPatch, comp.Milage]
	foragerFilterLoadPPP *ecs.Filter6[comp.Activity, comp.KnownPatch, comp.Milage, comp.NectarLoad, comp.EtoxLoad, comp.PPPExpo]
	foragerFilterSimple  *ecs.Filter2[comp.Activity, comp.KnownPatch]
	patchFilter          *ecs.Filter2[comp.Resource, comp.PatchProperties]
	patchUpdateFilter    *ecs.Filter7[comp.PatchProperties, comp.PatchDistance, comp.Resource, comp.HandlingTime, comp.Trip, comp.Mortality, comp.Dance]

	maxHoneyStore float64
}

func (s *ForagingEtox) Initialize(w *ecs.World) {
	s.foragerParams = ecs.GetResource[params.Foragers](w)
	s.forageParams = ecs.GetResource[params.Foraging](w)
	s.handlingTimeParams = ecs.GetResource[params.HandlingTime](w)
	s.danceParams = ecs.GetResource[params.Dance](w)
	s.energyParams = ecs.GetResource[params.EnergyContent](w)
	s.storeParams = ecs.GetResource[params.Stores](w)
	s.nursingParams = ecs.GetResource[params.Nursing](w)

	s.etox = ecs.GetResource[params.PPPApplication](w)
	s.toxic = ecs.GetResource[params.PPPToxicity](w)

	s.foragingStats = ecs.GetResource[globals.ForagingStats](w)
	s.foragePeriod = ecs.GetResource[globals.ForagingPeriod](w)
	s.stores = ecs.GetResource[globals.Stores](w)
	s.storesEtox = ecs.GetResource[globals.StoragesEtox](w)
	s.pppfate = ecs.GetResource[globals.PPPFate](w)
	s.pop = ecs.GetResource[globals.PopulationStats](w)
	s.newCohorts = ecs.GetResource[globals.NewCohorts](w)
	s.aff = ecs.GetResource[globals.AgeFirstForaging](w)
	s.factory = ecs.GetResource[globals.ForagerFactory](w)

	s.activityFilter = s.activityFilter.New(w)
	s.ageFilter = s.ageFilter.New(w)
	s.loadFilter = s.loadFilter.New(w)
	s.loadexpoFilter = s.loadexpoFilter.New(w)
	s.foragerFilter = s.foragerFilter.New(w)
	s.foragerFilterLoadPPP = s.foragerFilterLoadPPP.New(w)
	s.foragerFilterSimple = s.foragerFilterSimple.New(w)
	s.patchFilter = s.patchFilter.New(w)
	s.patchUpdateFilter = s.patchUpdateFilter.New(w)

	s.ageMapper = s.ageMapper.New(w)
	s.patchResourceMapper = s.patchResourceMapper.New(w)
	s.patchVisitsMapper = s.patchVisitsMapper.New(w)
	s.patchDanceMapper = s.patchDanceMapper.New(w)
	s.patchTripMapper = s.patchTripMapper.New(w)
	s.patchMortalityMapper = s.patchMortalityMapper.New(w)
	s.patchConfigMapper = s.patchConfigMapper.New(w)
	s.patchConfigMapperEtox = s.patchConfigMapperEtox.New(w)
	s.foragerMapper = s.foragerMapper.New(w)
	s.foragerLoadPPPMapper = s.foragerLoadPPPMapper.New(w)
	s.pppExpoAdder = s.pppExpoAdder.New(w)

	storeParams := ecs.GetResource[params.Stores](w)
	energyParams := ecs.GetResource[params.EnergyContent](w)

	s.maxHoneyStore = storeParams.MaxHoneyStoreKg * 1000.0 * energyParams.Honey
	s.rng = rand.New(ecs.GetResource[resource.Rand](w))
	s.time = ecs.GetResource[resource.Tick](w)
}

func (s *ForagingEtox) Update(w *ecs.World) {

	if s.newCohorts.Foragers > 0 {
		s.newForagers(w) // here the foragers get initialized now; mimics BEEHAVE exactly.
	}

	if s.foragePeriod.SecondsToday <= 0 ||
		(s.stores.Honey >= 0.95*s.maxHoneyStore && s.stores.Pollen >= s.stores.IdealPollen) {
		return
	}

	query := s.foragerFilter.Query()
	for query.Next() {
		_, _, milage := query.Get()
		milage.Today = 0
	}

	hangAroundDuration := s.forageParams.SearchLength / s.foragerParams.FlightVelocity
	forageProb := s.calcForagingProb()

	// TODO: Lazy winter bees.
	round := 0
	totalDuration := 0.0
	for {
		duration, foragers := s.foragingRound(w, forageProb)
		meanDuration := 0.0
		if foragers > 0 {
			meanDuration = duration / float64(foragers)
		} else {
			meanDuration = hangAroundDuration
		}
		totalDuration += meanDuration

		if totalDuration >= float64(s.foragePeriod.SecondsToday) {
			break
		}

		round++
	}
	query = s.foragerFilter.Query()
	for query.Next() {
		act, _, _ := query.Get()
		act.Current = activity.Resting
	}
}

func (s *ForagingEtox) Finalize(w *ecs.World) {}

func (s *ForagingEtox) newForagers(w *ecs.World) {
	if s.newCohorts.Foragers > 0 {
		s.factory.CreateSquadrons(s.newCohorts.Foragers, int(s.time.Tick)-s.aff.Aff)
	}
	s.newCohorts.Foragers = 0

	agequery := s.ageFilter.Without(ecs.C[comp.EtoxLoad]()).Query()
	for agequery.Next() {
		s.toAdd = append(s.toAdd, agequery.Entity())
	}

	for _, e := range s.toAdd {
		// adding etox components to the newly initialized forager entities
		s.pppExpoAdder.Add(e, &comp.PPPExpo{OralDose: 0., ContactDose: 0., RdmSurvivalContact: s.rng.Float64(), RdmSurvivalOral: s.rng.Float64()}, &comp.EtoxLoad{PPPLoad: 0.})
	}
	s.toAdd = s.toAdd[:0]
}

func (s *ForagingEtox) calcForagingProb() float64 {
	if s.stores.Pollen/s.stores.IdealPollen > 0.5 && s.stores.Honey/s.stores.DecentHoney > 1 {
		return 0
	}
	prob := s.forageParams.ProbBase
	if s.stores.Pollen/s.stores.IdealPollen < 0.2 || s.stores.Honey/s.stores.DecentHoney < 0.5 {
		prob = s.forageParams.ProbHigh
	}
	if s.stores.Honey/s.stores.DecentHoney < 0.2 {
		prob = s.forageParams.ProbEmergency
	}
	return prob
}

func (s *ForagingEtox) foragingRound(w *ecs.World, forageProb float64) (duration float64, foragers int) {
	probCollectPollen := (1.0 - s.stores.Pollen/s.stores.IdealPollen) * s.danceParams.MaxProportionPollenForagers

	s.stores.DecentHoney = math.Max(float64(s.pop.WorkersInHive+s.pop.WorkersForagers), 1) * s.storeParams.DecentHoneyPerWorker * s.energyParams.Honey // added this here, because Netlogo recalculates this in foragingRound and a countingproc happened since last calc.
	if s.stores.Honey/s.stores.DecentHoney < 0.5 {
		probCollectPollen *= s.stores.Honey / s.stores.DecentHoney
	}

	s.updatePatches(w)
	s.decisions(w, forageProb, probCollectPollen)
	s.searching(w)
	s.collecting(w)
	duration, foragers = s.flightCost(w)
	s.mortality(w)
	s.dancing(w)
	s.unloading(w)
	s.countForagers(w)
	return
}

func (s *ForagingEtox) updatePatches(w *ecs.World) {
	query := s.patchUpdateFilter.Query()
	for query.Next() {
		conf, dist, r, ht, trip, mort, dance := query.Get()

		if s.handlingTimeParams.ConstantHandlingTime {
			ht.Pollen = s.handlingTimeParams.PollenGathering
			ht.Nectar = s.handlingTimeParams.NectarGathering
		} else {
			ht.Pollen = s.handlingTimeParams.PollenGathering * r.MaxPollen / r.Pollen
			ht.Nectar = s.handlingTimeParams.NectarGathering * r.MaxNectar / r.Nectar
		}

		trip.CostNectar = (2 * dist.DistToColony * s.foragerParams.FlightCostPerM) +
			(s.foragerParams.FlightCostPerM * ht.Nectar *
				s.foragerParams.FlightVelocity * s.forageParams.EnergyOnFlower) // [kJ] = [m*kJ/m + kJ/m * s * m/s]

		trip.CostPollen = (2 * dist.DistToColony * s.foragerParams.FlightCostPerM) +
			(s.foragerParams.FlightCostPerM * ht.Pollen *
				s.foragerParams.FlightVelocity * s.forageParams.EnergyOnFlower) // [kJ] = [m*kJ/m + kJ/m * s * m/s]

		r.EnergyEfficiency = (conf.NectarConcentration*s.foragerParams.NectarLoad*s.energyParams.Sucrose - trip.CostNectar) / trip.CostNectar

		trip.DurationNectar = 2*dist.DistToColony/s.foragerParams.FlightVelocity + ht.Nectar
		trip.DurationPollen = 2*dist.DistToColony/s.foragerParams.FlightVelocity + ht.Pollen

		mort.Nectar = 1.0 - (math.Pow(1.0-s.forageParams.MortalityPerSec, trip.DurationNectar))
		mort.Pollen = 1.0 - (math.Pow(1.0-s.forageParams.MortalityPerSec, trip.DurationPollen))

		circ := r.EnergyEfficiency*s.danceParams.Slope + s.danceParams.Intercept
		dance.Circuits = util.Clamp(circ, 0, float64(s.danceParams.MaxCircuits))
	}
}

func (s *ForagingEtox) decisions(w *ecs.World, probForage, probCollectPollen float64) {
	query := s.foragerFilter.Query()
	for query.Next() {
		act, patch, milage := query.Get()

		if act.Current != activity.Recruited {
			act.PollenForager = s.rng.Float64() < probCollectPollen
		}

		if act.Current != activity.Recruited &&
			act.Current != activity.Resting &&
			act.Current != activity.Lazy {
			if s.rng.Float64() < s.forageParams.StopProbability {
				act.Current = activity.Resting
			}
		}

		if !act.PollenForager && !patch.Nectar.IsZero() {
			res := s.patchResourceMapper.Get(patch.Nectar)
			if s.rng.Float64() < 1.0/res.EnergyEfficiency &&
				s.rng.Float64() < s.stores.Honey/s.stores.DecentHoney {

				patch.Nectar = ecs.Entity{}
				if act.Current != activity.Resting && act.Current != activity.Lazy {
					act.Current = activity.Searching
				}
			}
		}

		if !patch.Pollen.IsZero() && act.PollenForager {
			trip := s.patchTripMapper.Get(patch.Pollen)
			if s.rng.Float64() < 1-math.Pow(1-s.forageParams.AbandonPollenPerSec, trip.DurationPollen) {
				patch.Nectar = ecs.Entity{}
				if act.Current != activity.Resting && act.Current != activity.Lazy {
					act.Current = activity.Searching
				}
			}
		}

		if act.Current == activity.Resting {
			if s.rng.Float64() < probForage {
				if act.PollenForager {
					if patch.Pollen.IsZero() {
						act.Current = activity.Searching
					} else {
						act.Current = activity.Experienced
					}
				} else {
					if patch.Nectar.IsZero() {
						act.Current = activity.Searching
					} else {
						act.Current = activity.Experienced
					}
				}
			}
		}

		if milage.Today > float32(s.foragerParams.MaxKmPerDay) {
			act.Current = activity.Resting
		}
	}
}

func (s *ForagingEtox) searching(w *ecs.World) {
	cumProb := 0.0
	nonDetectionProb := 1.0

	// TODO: water foraging search here, postponed because module seems to be rather irrelevant

	sz := float64(s.foragerParams.SquadronSize)
	patchQuery := s.patchFilter.Query()
	for patchQuery.Next() {
		res, conf := patchQuery.Get()
		hasNectar := res.Nectar >= s.foragerParams.NectarLoad*sz
		hasPollen := res.Pollen >= s.foragerParams.PollenLoad*sz
		if !hasNectar && !hasPollen {
			continue
		}
		s.patches = append(s.patches, patchCandidateEtox{
			Patch:       patchQuery.Entity(),
			Probability: conf.DetectionProbability,
			HasNectar:   hasNectar,
			HasPollen:   hasPollen,
		})

		cumProb += conf.DetectionProbability
		nonDetectionProb *= 1.0 - conf.DetectionProbability
	}
	detectionProb := 1.0 - nonDetectionProb

	// decoupled reruits from searchers and implemented shuffling via two separate sclices to imitate BEEHAVE more closely.
	activityQuery := s.activityFilter.Query()
	for activityQuery.Next() {
		act := activityQuery.Get()
		if act.Current == activity.Searching {
			s.searches = append(s.searches, activityQuery.Entity())
		} else if act.Current == activity.Recruited {
			s.recruits = append(s.recruits, activityQuery.Entity())
		}
	}

	s.rng.Shuffle(len(s.searches), func(i, j int) { s.searches[i], s.searches[j] = s.searches[j], s.searches[i] })
	for _, e := range s.searches {
		act, patch := s.foragerMapper.Get(e)

		if s.rng.Float64() >= detectionProb {
			continue
		}
		p := s.rng.Float64() * cumProb
		cum := 0.0
		var selected patchCandidateEtox
		for _, pch := range s.patches {
			cum += pch.Probability
			if cum >= p {
				selected = pch
				break
			}
		}
		if act.PollenForager {
			if selected.HasPollen {
				patch.Pollen = selected.Patch
				act.Current = activity.BringPollen
				res, vis := s.patchVisitsMapper.Get(selected.Patch)
				res.Pollen -= s.foragerParams.PollenLoad * sz
				vis.Pollen += s.foragerParams.SquadronSize
			} else {
				patch.Pollen = ecs.Entity{}
			}
		} else {
			if selected.HasNectar {
				patch.Nectar = selected.Patch
				act.Current = activity.BringNectar
				res, vis := s.patchVisitsMapper.Get(selected.Patch)
				res.Nectar -= s.foragerParams.NectarLoad * sz
				vis.Nectar += s.foragerParams.SquadronSize
			} else {
				patch.Nectar = ecs.Entity{}
			}
		}
	}

	s.rng.Shuffle(len(s.recruits), func(i, j int) { s.recruits[i], s.recruits[j] = s.recruits[j], s.recruits[i] })
	for _, e := range s.recruits {
		act, patch := s.foragerMapper.Get(e)

		if !act.PollenForager && !patch.Nectar.IsZero() {
			success := false

			if s.rng.Float64() < s.danceParams.FindProbability {
				res, vis := s.patchVisitsMapper.Get(patch.Nectar)
				if res.Nectar >= s.foragerParams.NectarLoad*sz {
					res.Nectar -= s.foragerParams.NectarLoad * sz
					vis.Nectar += s.foragerParams.SquadronSize
					act.Current = activity.BringNectar
					success = true
				}
			}
			if !success {
				act.Current = activity.Searching
				patch.Nectar = ecs.Entity{}
			}
		}

		if act.PollenForager && !patch.Pollen.IsZero() {
			success := false
			if s.rng.Float64() < s.danceParams.FindProbability {
				res, vis := s.patchVisitsMapper.Get(patch.Pollen)
				if res.Pollen >= s.foragerParams.PollenLoad*sz {
					res.Pollen -= s.foragerParams.PollenLoad * sz
					vis.Pollen += s.foragerParams.SquadronSize
					act.Current = activity.BringPollen
					success = true
				}
			}
			if !success {
				act.Current = activity.Searching
				patch.Pollen = ecs.Entity{}
			}
		}
	}
	//s.foragingStats.TotalSearches = len(s.foragerShuffle)
	s.patches = s.patches[:0]
	s.searches = s.searches[:0]
	s.recruits = s.recruits[:0]
}

func (s *ForagingEtox) collecting(w *ecs.World) {
	sz := float64(s.foragerParams.SquadronSize)

	// TODO: water collecting here, postponed because water foraging seems basically irrelevant overall
	activityQuery := s.activityFilter.Query()
	for activityQuery.Next() {
		act := activityQuery.Get()
		if act.Current == activity.Experienced || act.Current == activity.BringPollen || act.Current == activity.BringNectar {
			s.foragerShuffle = append(s.foragerShuffle, activityQuery.Entity())
		}
	}
	s.rng.Shuffle(len(s.foragerShuffle), func(i, j int) { s.foragerShuffle[i], s.foragerShuffle[j] = s.foragerShuffle[j], s.foragerShuffle[i] })

	for _, e := range s.foragerShuffle {
		act, patch, milage, load, PPPload, PPPexpo := s.foragerLoadPPPMapper.Get(e)

		if act.Current == activity.Experienced {
			if act.PollenForager {
				if patch.Pollen.IsZero() {
					act.Current = activity.Resting
				} else {
					res, vis := s.patchVisitsMapper.Get(patch.Pollen)
					if res.Pollen >= s.foragerParams.PollenLoad*sz {
						act.Current = activity.BringPollen
						res.Pollen -= s.foragerParams.PollenLoad * sz
						vis.Pollen += s.foragerParams.SquadronSize
					} else {
						act.Current = activity.Searching
						patch.Pollen = ecs.Entity{}
					}
				}
			} else {
				if patch.Nectar.IsZero() {
					act.Current = activity.Resting
				} else {
					res, vis := s.patchVisitsMapper.Get(patch.Nectar)
					if res.Nectar >= s.foragerParams.NectarLoad*sz {
						act.Current = activity.BringNectar
						res.Nectar -= s.foragerParams.NectarLoad * sz
						vis.Nectar += s.foragerParams.SquadronSize
					} else {
						act.Current = activity.Searching
						patch.Nectar = ecs.Entity{}
					}
				}
			}
		}

		if act.Current == activity.BringNectar {

			conf, etoxprops, trip := s.patchConfigMapperEtox.Get(patch.Nectar)
			load.Energy = conf.NectarConcentration * s.foragerParams.NectarLoad * s.energyParams.Sucrose // --> kJ
			dist := trip.CostNectar / (s.foragerParams.FlightCostPerM * 1000)
			milage.Today += float32(dist)
			milage.Total += float32(dist)

			// exposure from nectar foraging
			PPPload.PPPLoad = load.Energy * etoxprops.PPPconcentrationNectar // kJ * mug/kJ = mug / load
			PPPexpo.OralDose += PPPload.PPPLoad * s.toxic.HSuptake

			// pppfate is simply used to create a ppp mass balance to analyze and debug
			s.pppfate.PPPforagersImmediate += PPPload.PPPLoad * s.toxic.HSuptake * float64(s.foragerParams.SquadronSize)
			s.pppfate.PPPforagersTotal += PPPload.PPPLoad * s.toxic.HSuptake * float64(s.foragerParams.SquadronSize)
			s.pppfate.TotalPPPforaged += PPPload.PPPLoad * float64(s.foragerParams.SquadronSize)

			PPPload.PPPLoad -= PPPload.PPPLoad * s.toxic.HSuptake

			if s.etox.AppDay == int(s.time.Tick) || !s.etox.ContactExposureOneDay {
				if PPPexpo.ContactDose > 0 {
					if s.etox.ContactSum {
						PPPexpo.ContactDose += etoxprops.PPPcontactDose
					} else {
						PPPexpo.ContactDose = (PPPexpo.ContactDose + etoxprops.PPPcontactDose) / 2
					}
				} else {
					PPPexpo.ContactDose += etoxprops.PPPcontactDose
				}
			}
		}

		if act.Current == activity.BringPollen {

			_, etoxprops, trip := s.patchConfigMapperEtox.Get(patch.Pollen)
			dist := trip.CostPollen / (s.foragerParams.FlightCostPerM * 1000)
			milage.Today += float32(dist)
			milage.Total += float32(dist)

			// exposure from pollen foraging
			PPPload.PPPLoad = s.foragerParams.PollenLoad * etoxprops.PPPconcentrationPollen // g * mug/g = mug / load
			s.pppfate.TotalPPPforaged += PPPload.PPPLoad * float64(s.foragerParams.SquadronSize)

			if s.etox.AppDay == int(s.time.Tick) || !s.etox.ContactExposureOneDay {
				if PPPexpo.ContactDose > 0 {
					if s.etox.ContactSum {
						PPPexpo.ContactDose += etoxprops.PPPcontactDose
					} else {
						PPPexpo.ContactDose = (PPPexpo.ContactDose + etoxprops.PPPcontactDose) / 2
					}
				} else {
					PPPexpo.ContactDose += etoxprops.PPPcontactDose
				}
			}
		}
	}
	s.foragerShuffle = s.foragerShuffle[:0]
}

func (s *ForagingEtox) flightCost(w *ecs.World) (duration float64, foragers int) {
	duration = 0.0
	foragers = 0

	// TODO: flightCost for water foraging here, postponed because of a lack of relevance

	activityQuery := s.activityFilter.Query()
	for activityQuery.Next() {
		act := activityQuery.Get()
		if act.Current == activity.Searching || act.Current == activity.BringPollen || act.Current == activity.BringNectar {
			s.foragerShuffle = append(s.foragerShuffle, activityQuery.Entity())
		}
	}
	s.rng.Shuffle(len(s.foragerShuffle), func(i, j int) { s.foragerShuffle[i], s.foragerShuffle[j] = s.foragerShuffle[j], s.foragerShuffle[i] })

	for _, e := range s.foragerShuffle {
		act, patch, milage, _, _, ppp := s.foragerLoadPPPMapper.Get(e)

		if act.Current == activity.Searching {
			dist := s.forageParams.SearchLength / 1000.0
			milage.Today += float32(dist)
			milage.Total += float32(dist)

			en := s.forageParams.SearchLength * s.foragerParams.FlightCostPerM
			s.stores.Honey -= en * float64(s.foragerParams.SquadronSize)

			flightintake := s.FeedOnHoneyStores(w, en*float64(s.foragerParams.SquadronSize), float64(s.foragerParams.SquadronSize))

			s.pppfate.PPPforagersinHive += flightintake * float64(s.foragerParams.SquadronSize)
			s.pppfate.PPPforagersTotal += flightintake * float64(s.foragerParams.SquadronSize)
			ppp.OralDose += flightintake

			duration += s.forageParams.SearchLength / s.foragerParams.FlightVelocity
			foragers++
		} else if act.Current == activity.BringNectar || act.Current == activity.BringPollen {
			en := 0.0
			if act.PollenForager {
				trip := s.patchTripMapper.Get(patch.Pollen)
				duration += trip.DurationPollen + s.handlingTimeParams.PollenUnloading
				en = trip.CostPollen
			} else {
				trip := s.patchTripMapper.Get(patch.Nectar)
				duration += trip.DurationNectar + s.handlingTimeParams.NectarUnloading
				en = trip.CostNectar
			}
			s.stores.Honey -= en * float64(s.foragerParams.SquadronSize)

			flightintake := s.FeedOnHoneyStores(w, en*float64(s.foragerParams.SquadronSize), float64(s.foragerParams.SquadronSize))

			s.pppfate.PPPforagersinHive += flightintake * float64(s.foragerParams.SquadronSize)
			s.pppfate.PPPforagersTotal += flightintake * float64(s.foragerParams.SquadronSize)
			ppp.OralDose += flightintake

			foragers++
		}
	}
	s.foragerShuffle = s.foragerShuffle[:0]

	return
}

func (s *ForagingEtox) mortality(w *ecs.World) {
	searchDuration := s.forageParams.SearchLength / s.foragerParams.FlightVelocity

	// TODO: mortality for water foragers, postponed ..

	foragerQuery := s.foragerFilterLoadPPP.Query()
	for foragerQuery.Next() {
		act, patch, _, _, PPPload, PPPexpo := foragerQuery.Get()

		// Acute toxicity during flight
		lethaldose := false
		if s.etox.ForagerImmediateMortality { // always false for now; might as well be deactivated
			if PPPexpo.RdmSurvivalOral < 1-(1/(1+math.Pow(PPPexpo.OralDose/s.toxic.ForagerOralLD50, s.toxic.ForagerOralSlope))) {
				lethaldose = true
			}
			if PPPexpo.RdmSurvivalContact < 1-(1/(1+math.Pow(PPPexpo.ContactDose/s.toxic.ForagerContactLD50, s.toxic.ForagerContactSlope))) {
				lethaldose = true
			}
		}

		if lethaldose {
			s.toRemove = append(s.toRemove, foragerQuery.Entity())
		} else if act.Current == activity.Searching {
			if s.rng.Float64() < 1-math.Pow(1-s.forageParams.MortalityPerSec, searchDuration) {
				s.toRemove = append(s.toRemove, foragerQuery.Entity())
			}
		} else if act.Current == activity.BringNectar {
			m := s.patchMortalityMapper.Get(patch.Nectar)
			if s.rng.Float64() < m.Nectar || lethaldose {
				s.toRemove = append(s.toRemove, foragerQuery.Entity())
				s.pppfate.ForagerDiedInFlight += PPPload.PPPLoad * float64(s.foragerParams.SquadronSize)
			}
		} else if act.Current == activity.BringPollen {
			m := s.patchMortalityMapper.Get(patch.Pollen)
			if s.rng.Float64() < m.Pollen || lethaldose {
				s.toRemove = append(s.toRemove, foragerQuery.Entity())
				s.pppfate.ForagerDiedInFlight += PPPload.PPPLoad * float64(s.foragerParams.SquadronSize)
			}
		}
	}

	for _, e := range s.toRemove {
		w.RemoveEntity(e)
	}
	s.toRemove = s.toRemove[:0]
}

func (s *ForagingEtox) dancing(w *ecs.World) {
	activityQuery := s.activityFilter.Query()
	for activityQuery.Next() {
		act := activityQuery.Get()

		if act.Current == activity.Resting {
			s.resting = append(s.resting, activityQuery.Entity())
		} else if act.Current == activity.BringNectar || act.Current == activity.BringPollen {
			s.dances = append(s.dances, activityQuery.Entity())
		}
	}
	s.rng.Shuffle(len(s.resting), func(i, j int) { s.resting[i], s.resting[j] = s.resting[j], s.resting[i] })
	s.rng.Shuffle(len(s.dances), func(i, j int) { s.dances[i], s.dances[j] = s.dances[j], s.dances[i] })

	for _, e := range s.dances {
		act, patch := s.foragerMapper.Get(e)

		if act.Current != activity.BringNectar && act.Current != activity.BringPollen {
			continue
		}

		if act.Current == activity.BringNectar {
			patchRes, dance := s.patchDanceMapper.Get(patch.Nectar)
			danceEEF := patchRes.EnergyEfficiency

			rPoisson := distuv.Poisson{
				Src:    &util.RandWrapper{Src: s.rng},
				Lambda: dance.Circuits * 0.05,
			}
			danceFollowers := int(rPoisson.Rand())

			if danceFollowers == 0 {
				continue
			}
			if len(s.resting) < danceFollowers {
				continue
			}

			for i := 0; i < danceFollowers; i++ {
				follower := s.resting[len(s.resting)-1]
				fAct, fPatch := s.foragerMapper.Get(follower)

				if fPatch.Nectar.IsZero() {
					fPatch.Nectar = patch.Nectar
					fAct.Current = activity.Recruited
					fAct.PollenForager = false
				} else {
					knownRes := s.patchResourceMapper.Get(fPatch.Nectar)
					if danceEEF > knownRes.EnergyEfficiency {
						fPatch.Nectar = patch.Nectar
						fAct.Current = activity.Recruited
						fAct.PollenForager = false
					} else {
						// TODO: really? was resting before!
						fAct.Current = activity.Experienced
					}
				}

				s.resting = s.resting[:len(s.resting)-1]
			}
		}

		if act.Current == activity.BringPollen {
			trip := s.patchTripMapper.Get(patch.Pollen)
			danceTripDuration := trip.DurationPollen

			danceFollowers := s.danceParams.PollenDanceFollowers

			if len(s.resting) < danceFollowers {
				continue
			}

			for i := 0; i < danceFollowers; i++ {
				follower := s.resting[len(s.resting)-1]
				fAct, fPatch := s.foragerMapper.Get(follower)

				if fPatch.Pollen.IsZero() {
					fPatch.Pollen = patch.Pollen
					fAct.Current = activity.Recruited
					fAct.PollenForager = true
				} else {
					knownTrip := s.patchTripMapper.Get(fPatch.Pollen)
					if danceTripDuration < knownTrip.DurationPollen {
						fPatch.Pollen = patch.Pollen
						fAct.Current = activity.Recruited
						fAct.PollenForager = true
					} else {
						// TODO: really? was resting before!
						fAct.Current = activity.Experienced
					}
				}

				s.resting = s.resting[:len(s.resting)-1]
			}
		}
	}

	s.resting = s.resting[:0]
	s.dances = s.dances[:0]
}

func (s *ForagingEtox) unloading(w *ecs.World) {

	// TODO: water unloading, postponed ...

	query := s.loadexpoFilter.Query()
	for query.Next() {
		act, load, PPPload, ppp := query.Get()
		if act.Current == activity.BringNectar {

			s.stores.Honey += load.Energy * float64(s.foragerParams.SquadronSize)

			s.storesEtox.EtoxHoneyConcentration[0] = ((s.storesEtox.EtoxHoneyConcentration[0] * s.storesEtox.EtoxHoneyEnergy[0]) +
				(PPPload.PPPLoad * (1 - s.toxic.HSuptake) * float64(s.foragerParams.SquadronSize))) /
				(s.storesEtox.EtoxHoneyEnergy[0] + (load.Energy * float64(s.foragerParams.SquadronSize)))
			// HSuptake actually gets applied a second time in here; it already got applied to PPPload when the foragers took up the load; the lost fraction then got added to the foragers OralDose
			// here PPPLoad loses another 10% (in total 19% are "lost" to the forager), but these 10% just dissipate. There is no addition to foragers OralDose, 9% of total pesticide taken in via nectarforaging is thus lost in the model without the fix below
			// BEEHAVE_ecotox ODD and various suppl. resources do not talk about HSuptake anywhere sadly
			if s.etox.HSUfix {
				ppp.OralDose += PPPload.PPPLoad * s.toxic.HSuptake
				s.pppfate.PPPforagersImmediate += PPPload.PPPLoad * s.toxic.HSuptake * float64(s.foragerParams.SquadronSize)
				s.pppfate.PPPforagersTotal += PPPload.PPPLoad * s.toxic.HSuptake * float64(s.foragerParams.SquadronSize)
			}
			s.pppfate.PPPhoneyStores += PPPload.PPPLoad * (1 - s.toxic.HSuptake) * float64(s.foragerParams.SquadronSize)

			s.storesEtox.EtoxHoneyEnergy[0] += load.Energy * float64(s.foragerParams.SquadronSize)
			if s.stores.Honey > s.maxHoneyStore {
				s.stores.Honey = s.maxHoneyStore
				s.storesEtox.EtoxHoneyEnergy[0] = s.maxHoneyStore - (s.storesEtox.EtoxHoneyEnergy[1] + s.storesEtox.EtoxHoneyEnergy[2] + s.storesEtox.EtoxHoneyEnergy[3] + s.storesEtox.EtoxHoneyEnergy[4] + s.storesEtox.EtoxHoneyEnergy[5])
			}

			load.Energy = 0.
			PPPload.PPPLoad = 0.
			act.Current = activity.Experienced
		} else if act.Current == activity.BringPollen {
			s.storesEtox.PPPInHivePollenConc = ((s.storesEtox.PPPInHivePollenConc * s.stores.Pollen) + (PPPload.PPPLoad * float64(s.foragerParams.SquadronSize))) / (s.stores.Pollen + s.foragerParams.PollenLoad*float64(s.foragerParams.SquadronSize)) // may need to readjust

			s.pppfate.PPPpollenStores += PPPload.PPPLoad * float64(s.foragerParams.SquadronSize)

			s.stores.Pollen += s.foragerParams.PollenLoad * float64(s.foragerParams.SquadronSize)
			PPPload.PPPLoad = 0.
			act.Current = activity.Experienced
		}
	}
}

func (s *ForagingEtox) countForagers(w *ecs.World) {
	sz := s.foragerParams.SquadronSize
	round := globals.ForagingRound{}

	query := s.activityFilter.Query()
	for query.Next() {
		act := query.Get()

		switch act.Current {
		case activity.Lazy:
			round.Lazy += sz
		case activity.Resting:
			round.Resting += sz
		case activity.Searching:
			round.Searching += sz
		case activity.Recruited:
			round.Recruited += sz
		case activity.Experienced:
			if act.PollenForager {
				round.Pollen += sz
			} else {
				round.Nectar += sz
			}
		default:
			panic(fmt.Sprintf("forager activity %d invalid at the end of a foraging round", act.Current))
		}
	}

	s.foragingStats.Rounds = append(s.foragingStats.Rounds, round)
}

type patchCandidateEtox struct {
	Patch       ecs.Entity
	Probability float64
	HasNectar   bool
	HasPollen   bool

	HasWater bool
}

// copy from etox_storages_consumption
func (s *ForagingEtox) FeedOnHoneyStores(w *ecs.World, cons float64, number float64) (oralDose float64) {
	oralDose = 0.

	// deflated the FeedOnHoneyStores function by introducing a slice for the honey stores and their PPP concentrations
	// and simply iterating over this slice and breaking it in the process if consumption goal is reached.
	for i, _ := range s.storesEtox.EtoxHoneyEnergy {
		if cons < s.storesEtox.EtoxHoneyEnergy[i] {
			oralDose += cons * s.storesEtox.EtoxHoneyConcentration[i] / number
			s.storesEtox.EtoxHoneyEnergy[i] -= cons
			break
		} else {
			oralDose += s.storesEtox.EtoxHoneyEnergy[i] * s.storesEtox.EtoxHoneyConcentration[i] / number
			cons -= s.storesEtox.EtoxHoneyEnergy[i]
			s.storesEtox.EtoxHoneyEnergy[i] = 0
		}
	}
	return
}
