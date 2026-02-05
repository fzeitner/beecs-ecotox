package sys

// regulates the compartimentalized storages of the _ecotox addition
// and classic BEEHAVE_ecotox calculation of exposure
// updates concentrations of PPP in nectar
// corresponding process in NetLogo: TupdateInternalExposureNectar_ETOX
// all cohorts work with a mean dose per cohort that gets calculated based on number of individuals in that cohort and their consumption rates

import (
	"math"
	"math/rand/v2"

	"github.com/mlange-42/ark-tools/resource"
	"github.com/mlange-42/ark/ecs"
	"github.com/mlange-42/beecs/comp"
	"github.com/mlange-42/beecs/globals"
	"github.com/mlange-42/beecs/params"
)

type EtoxStorages struct {
	needs          *params.HoneyNeeds
	needsPollen    *params.PollenNeeds
	workerDev      *params.WorkerDevelopment
	oldNurseParams *params.Nursing
	energyParams   *params.EnergyContent
	storesParams   *params.Stores
	foragerParams  *params.Foragers
	etox           *params.PPPApplication
	toxic          *params.PPPToxicity

	beecsStores *globals.Stores
	stores      *globals.StoragesEtox
	pppFate     *globals.PPPFate
	pop         *globals.PopulationStats
	etoxStats   *globals.PopulationStatsEtox
	inHive      *globals.InHive
	inHiveEtox  *globals.InHiveEtox
	Larvae      *globals.Larvae
	LarvaeEtox  *globals.LarvaeEtox
	cons        *globals.ConsumptionStats

	foragerExpoMapper     *ecs.Map1[comp.PPPExpo]
	foragerActivityMapper *ecs.Map1[comp.Activity]
	foragerFilter         *ecs.Filter1[comp.Age]
	foragerShuffle        []ecs.Entity

	rng *rand.Rand
}

func (s *EtoxStorages) Initialize(w *ecs.World) {
	s.needs = ecs.GetResource[params.HoneyNeeds](w)
	s.needsPollen = ecs.GetResource[params.PollenNeeds](w)
	s.workerDev = ecs.GetResource[params.WorkerDevelopment](w)
	s.oldNurseParams = ecs.GetResource[params.Nursing](w)
	s.energyParams = ecs.GetResource[params.EnergyContent](w)
	s.storesParams = ecs.GetResource[params.Stores](w)
	s.foragerParams = ecs.GetResource[params.Foragers](w)
	s.etox = ecs.GetResource[params.PPPApplication](w)
	s.toxic = ecs.GetResource[params.PPPToxicity](w)

	s.beecsStores = ecs.GetResource[globals.Stores](w)
	s.stores = ecs.GetResource[globals.StoragesEtox](w)
	s.pop = ecs.GetResource[globals.PopulationStats](w)
	s.etoxStats = ecs.GetResource[globals.PopulationStatsEtox](w)
	s.pppFate = ecs.GetResource[globals.PPPFate](w)
	s.inHive = ecs.GetResource[globals.InHive](w)
	s.inHiveEtox = ecs.GetResource[globals.InHiveEtox](w)
	s.Larvae = ecs.GetResource[globals.Larvae](w)
	s.LarvaeEtox = ecs.GetResource[globals.LarvaeEtox](w)
	s.cons = ecs.GetResource[globals.ConsumptionStats](w)

	s.foragerExpoMapper = s.foragerExpoMapper.New(w)
	s.foragerActivityMapper = s.foragerActivityMapper.New(w)
	s.foragerFilter = s.foragerFilter.New(w)

	s.rng = rand.New(ecs.GetResource[resource.Rand](w))
}

func (s *EtoxStorages) Update(w *ecs.World) {
	// initiate necessary variables
	h := 0.  // for tracking honey in between cohorts
	p := 0.  // for tracking pollen in between cohorts
	c := 0   // for tracking the amount of individuals in the cohorts
	num := 0 // for tracking number of total individuals within one caste
	s.etoxStats.CumDoseNurses = 0.

	consumed_pollen := 0. // tracker for total amount of pollen consumed in this subsystem
	consumed_honey := 0.  // tracker for total amount of honey consumed in this subsystem

	// Thermoregulation energy budget
	thermoRegBrood := (s.needs.WorkerNurse - s.needs.WorkerResting) / s.oldNurseParams.MaxBroodNurseRatio
	workerbaselineneed := s.needs.WorkerResting
	if s.etox.ReworkedThermoETOX {
		s.stores.EtoxEnergyThermo = float64(s.pop.TotalBrood) * thermoRegBrood / float64(s.pop.WorkersForagers+s.pop.WorkersInHive) // calculate how much honey each adult IHbee/forager would need to take in extra
		workerbaselineneed += s.stores.EtoxEnergyThermo
		s.stores.EtoxEnergyThermo = 0.
	} else {
		s.stores.EtoxEnergyThermo = float64(s.pop.TotalBrood) * thermoRegBrood * 0.001 * s.energyParams.Honey // or calculate the total necessary energy
	}

	// get values for some observing/debugging variables
	s.stores.Pollenconcbeforeeating = s.stores.PPPInHivePollenConc // used in debugging and as a helpful metric

	s.stores.Nectarconcbeforeeating = 0 // used in debugging and as a helpful metric
	for i := range s.stores.EtoxHoneyEnergy {
		if s.stores.EtoxHoneyEnergy[i] > 0 {
			s.stores.Nectarconcbeforeeating = s.stores.EtoxHoneyConcentration[i]
			break
		}
	}

	// foragers, pretty straigt forward and same for all model versions
	forquery := s.foragerFilter.Query()
	for forquery.Next() {
		s.foragerShuffle = append(s.foragerShuffle, forquery.Entity())
	}
	s.rng.Shuffle(len(s.foragerShuffle), func(i, j int) { s.foragerShuffle[i], s.foragerShuffle[j] = s.foragerShuffle[j], s.foragerShuffle[i] })
	forcount := len(s.foragerShuffle) * 100

	for _, e := range s.foragerShuffle {
		ppp := s.foragerExpoMapper.Get(e)
		ppp.OralDose += s.stores.PPPInHivePollenConc * s.needsPollen.Worker * 0.001
		s.pppFate.PPPforagersinHive += s.stores.PPPInHivePollenConc * s.needsPollen.Worker * 0.001 * float64(s.foragerParams.SquadronSize)
		s.pppFate.PPPforagersTotal += s.stores.PPPInHivePollenConc * s.needsPollen.Worker * 0.001 * float64(s.foragerParams.SquadronSize)

		ETOX_Consumed := workerbaselineneed * 0.001 * s.energyParams.Honey * float64(s.foragerParams.SquadronSize)
		ETOX_Consumed += s.stores.EtoxEnergyThermo
		s.stores.EtoxEnergyThermo = 0.

		intake := s.FeedOnHoneyStores(w, ETOX_Consumed, float64(s.foragerParams.SquadronSize))
		ppp.OralDose += intake

		s.pppFate.PPPforagersinHive += intake * float64(s.foragerParams.SquadronSize)
		s.pppFate.PPPforagersTotal += intake * float64(s.foragerParams.SquadronSize)

		consumed_pollen += s.needsPollen.Worker * float64(s.foragerParams.SquadronSize)
		consumed_honey += ETOX_Consumed
	}
	s.foragerShuffle = s.foragerShuffle[:0]

	// inhive bees, all cohorts work with a mean dose per cohort that gets calculated based on number of individuals in that cohort and their consumption rates
	s.etoxStats.CumDoseIHBees, c, h, p, num = s.CalcDosePerCohort(w, s.inHive.Workers, s.inHiveEtox.WorkerCohortDose, s.stores.EtoxEnergyThermo, workerbaselineneed, s.needsPollen.Worker, float64(1), float64(1))
	s.stores.EtoxEnergyThermo = 0.
	currentIHbees := num
	if s.pop.WorkersInHive > 0 {
		s.etoxStats.MeanDoseIHBees = s.etoxStats.CumDoseIHBees / float64(currentIHbees)
	} else {
		s.etoxStats.MeanDoseIHBees = 0.
	}
	s.etoxStats.NumberIHbeeCohorts = c

	consumed_honey += h
	consumed_pollen += p
	s.pppFate.PPPIHbees += s.etoxStats.CumDoseIHBees

	// inhive larvae, all cohorts work with a mean dose per cohort that gets calculated based on number of individuals in that cohort and their consumption rates
	// larvae exposure considers the nursebee-filtering effect
	s.etoxStats.CumDoseLarvae, _, h, p, num = s.CalcDosePerCohort(w, s.Larvae.Workers, s.LarvaeEtox.WorkerCohortDose, s.stores.EtoxEnergyThermo, (s.needs.WorkerLarvaTotal / float64(s.workerDev.LarvaeTime)), (s.needsPollen.WorkerLarvaTotal / float64(s.workerDev.LarvaeTime)), s.toxic.NursebeesNectar, s.toxic.NursebeesPollen)
	if s.pop.WorkerLarvae > 0 {
		s.etoxStats.MeanDoseLarvae = s.etoxStats.CumDoseLarvae / float64(num)
	} else {
		s.etoxStats.MeanDoseLarvae = 0.
	}

	consumed_honey += h
	consumed_pollen += p
	s.pppFate.PPPlarvae += s.etoxStats.CumDoseLarvae

	// inhive dronelarvae, all cohorts work with a mean dose per cohort that gets calculated based on number of individuals in that cohort and their consumption rates
	// larvae exposure considers the nursebee-filtering effect
	s.etoxStats.CumDoseDroneLarvae, _, h, p, num = s.CalcDosePerCohort(w, s.Larvae.Drones, s.LarvaeEtox.DroneCohortDose, s.stores.EtoxEnergyThermo, s.needs.DroneLarva, s.needsPollen.DroneLarva, s.toxic.NursebeesNectar, s.toxic.NursebeesPollen)
	if s.pop.DroneLarvae > 0 {
		s.etoxStats.MeanDoseDroneLarvae = s.etoxStats.CumDoseDroneLarvae / float64(num)
	} else {
		s.etoxStats.MeanDoseDroneLarvae = 0.
	}

	consumed_honey += h
	consumed_pollen += p
	s.pppFate.PPPdlarvae += s.etoxStats.CumDoseDroneLarvae

	if s.pop.WorkersInHive > 0 {
		s.etoxStats.CumDoseNurses = s.etoxStats.CumDoseIHBees + s.etoxStats.PPPNursebees
		s.etoxStats.MeanDoseNurses = s.etoxStats.CumDoseNurses / float64(currentIHbees)
	} else {
		s.etoxStats.CumDoseNurses = 0.
		s.etoxStats.MeanDoseNurses = 0.
	}
	s.pppFate.PPPNurses += s.etoxStats.PPPNursebees // this is the intake from Nursebeefactors of the old model version
	if s.etox.Nursebeefix && s.pop.WorkersInHive != 0 {
		s.addNurseExptoIHbees(w, s.etoxStats.PPPNursebees, float64(currentIHbees), s.inHive.Workers, s.inHiveEtox.WorkerCohortDose)
	}

	// inhive drones, all cohorts work with a mean dose per cohort that gets calculated based on number of individuals in that cohort and their consumption rates
	s.etoxStats.CumDoseDrones, _, h, p, num = s.CalcDosePerCohort(w, s.inHive.Drones, s.inHiveEtox.DroneCohortDose, s.stores.EtoxEnergyThermo, s.needs.Drone, s.needsPollen.Drone, float64(1), float64(1))
	if s.pop.DroneLarvae > 0 {
		s.etoxStats.MeanDoseDrones = s.etoxStats.CumDoseDrones / float64(num)
	} else {
		s.etoxStats.MeanDoseDrones = 0.
	}

	consumed_honey += h
	consumed_pollen += p
	s.pppFate.PPPdrones += s.etoxStats.CumDoseDrones

	if s.etox.DegradationHoney {
		s.DegradeHoney(w)
	}

	// leftovers from debugging
	_ = s.pop.DroneLarvae + s.pop.DronesInHive + s.pop.WorkerLarvae + s.pop.WorkersForagers + s.pop.WorkersInHive + forcount
	// checkpoint for bugfixing honey consumption in etox
	if math.Round(consumed_honey-(s.cons.HoneyDaily*0.001*s.energyParams.Honey)) != 0 || math.Round(consumed_pollen/1000.0-s.cons.PollenDaily) != 0 {
		panic("Fatal error in honey store dose calculations, model output will be wrong!")
	}

	s.ShiftHoney(w)

	s.stores.PPPpollenTotal = s.stores.PPPInHivePollenConc * s.beecsStores.Pollen
	s.stores.PPPhoneyTotal = s.calcPPPhoneytotal(w)
	s.stores.PPPTotal = s.stores.PPPpollenTotal + s.stores.PPPhoneyTotal
}

func (s *EtoxStorages) addNurseExptoIHbees(w *ecs.World, PPPnurses float64, NumIHbees float64, coh []int, dose []float64) {

	AddOralDose := PPPnurses / NumIHbees

	for i := range coh {
		if coh[i] != 0 {
			dose[i] += AddOralDose
			PPPnurses -= AddOralDose * float64(coh[i])
		}
	}
	if math.Round(PPPnurses) != 0 {
		panic("PPP should be 0 by now, there must be a bug somewhere!")
	}
}

func (s *EtoxStorages) calcPPPhoneytotal(w *ecs.World) (totalPPP float64) {
	totalPPP = 0.
	for i := range s.stores.EtoxHoneyEnergy {
		totalPPP += s.stores.EtoxHoneyEnergy[i] * s.stores.EtoxHoneyConcentration[i]
	}
	return
}

func (s *EtoxStorages) CalcDosePerCohort(w *ecs.World, coh []int, dose []float64, init_honeyenergy float64, honey_need float64, pollen_need float64, nursebeefactorHoney float64, nursebeefactorPollen float64) (CumDose float64, cohortcounter int, consumed float64, pconsumed float64, num int) {
	// this is the baseline version with the logic of the original BEEHAVE_ecotox function
	CumDose = 0.
	cohortcounter = 0
	consumed = 0.
	num = 0
	pconsumed = 0.

	order := rand.Perm(len(coh)) // randomize order to further emulate NetLogo ask function
	for _, i := range order {
		ETOX_PPPOralDose := 0.
		ETOX_Consumed_Honey := init_honeyenergy

		if coh[i] != 0 {
			init_honeyenergy = 0.
			cohortcounter++
			num += coh[i]

			ETOX_Consumed_Honey += honey_need * 0.001 * s.energyParams.Honey * float64(coh[i])
			ETOX_PPPOralDose += s.FeedOnHoneyStores(w, ETOX_Consumed_Honey, float64(coh[i])) // calculates the exposure from consumption of honey storage

			if s.etox.Nursebeefix {
				s.etoxStats.PPPNursebees += ETOX_PPPOralDose * (1 - nursebeefactorHoney) * float64(coh[i])
				s.etoxStats.PPPNursebees += s.stores.PPPInHivePollenConc * pollen_need * 0.001 * (1 - nursebeefactorPollen) * float64(coh[i])
			}
			ETOX_PPPOralDose = ETOX_PPPOralDose * nursebeefactorHoney
			ETOX_PPPOralDose += s.stores.PPPInHivePollenConc * pollen_need * 0.001 * nursebeefactorPollen // intake from pollen

			consumed += ETOX_Consumed_Honey
			pconsumed += pollen_need * float64(coh[i])

			dose[i] = ETOX_PPPOralDose
			CumDose += ETOX_PPPOralDose * float64(coh[i])
		} else {
			dose[i] = 0
		}
	}
	return
}

func (s *EtoxStorages) FeedOnHoneyStores(w *ecs.World, cons float64, number float64) (oralDose float64) {
	oralDose = 0.

	// deflated the FeedOnHoneyStores function by introducing a slice for the honey stores and their PPP concentrations
	// and simply iterating over this slice and breaking it in the process if consumption goal is reached.
	for i, _ := range s.stores.EtoxHoneyEnergy {
		if cons < s.stores.EtoxHoneyEnergy[i] {
			oralDose += cons * s.stores.EtoxHoneyConcentration[i] / number
			s.stores.EtoxHoneyEnergy[i] -= cons
			break
		} else {
			oralDose += s.stores.EtoxHoneyEnergy[i] * s.stores.EtoxHoneyConcentration[i] / number
			cons -= s.stores.EtoxHoneyEnergy[i]
			s.stores.EtoxHoneyEnergy[i] = 0
		}
	}
	return
}

func (s *EtoxStorages) DegradeHoney(w *ecs.World) {
	// if this ever is to be turned on PPPfate oberver should be introduced here if someone wanted to create a mass balance again
	DT50honey := s.etox.DT50honey
	for i := range s.stores.EtoxHoneyEnergy {
		s.stores.EtoxHoneyConcentration[i] = s.stores.EtoxHoneyConcentration[i] * math.Exp(-math.Log(2)/DT50honey) // Dissappearance of the pesticide in the honey following a single first-order kinetic
	}
}

func (s *EtoxStorages) ShiftHoney(w *ecs.World) {
	// shift from the 4 day old honey stores to capped first and combine their PPP concentrations to make place for the other stores
	combinedHoney := s.stores.EtoxHoneyEnergy[4] + s.stores.EtoxHoneyEnergy[5]

	if combinedHoney > 0 {
		s.stores.EtoxHoneyConcentration[5] = (s.stores.EtoxHoneyEnergy[5]*s.stores.EtoxHoneyConcentration[5] +
			s.stores.EtoxHoneyEnergy[4]*s.stores.EtoxHoneyConcentration[4]) / combinedHoney
		s.stores.EtoxHoneyEnergy[5] = combinedHoney
	}

	// shift honey and concentration of all other stores by one day
	for i := 4; i > 0; i-- {
		s.stores.EtoxHoneyEnergy[i] = s.stores.EtoxHoneyEnergy[i-1]
		s.stores.EtoxHoneyConcentration[i] = s.stores.EtoxHoneyConcentration[i-1]
	}
	// reset the freshest stores to 0
	s.stores.EtoxHoneyEnergy[0] = 0
	s.stores.EtoxHoneyConcentration[0] = 0

	// in the normal stores of capped etox stores are < 0 this means all honey has been used up. In this case everything gets reset and
	// the hive will die in the next timestep. This is only to make sure no bugs occur from negative stores in this case.
	if s.beecsStores.Honey <= 0 || s.stores.EtoxHoneyConcentration[5] < 0 {
		s.beecsStores.Honey = 0
		for i := range s.stores.EtoxHoneyEnergy {
			s.stores.EtoxHoneyEnergy[i] = 0
			s.stores.EtoxHoneyConcentration[i] = 0
		}
	}

	// check for any large deviations between the stores to make sure the subsystem works as intended
	sumEtoxStores := 0.
	for _, val := range s.stores.EtoxHoneyEnergy {
		sumEtoxStores += val
	}
	// adjusted this panic to 0.1% acceptable deviation from the honey store in each timestep; 0.1% deemed acceptable due to possibility of floating point errors on many occasions
	if math.Round((sumEtoxStores))*1.001 < math.Round(s.beecsStores.Honey) ||
		math.Round((sumEtoxStores))*0.999 > math.Round(s.beecsStores.Honey) {
		panic("Fatal error in honey store dose calculations, model output will be wrong!") // should debug why this triggers sometimes on very long simulation runs through small differences once there is time
	}

}

func (s *EtoxStorages) Finalize(w *ecs.World) {}
