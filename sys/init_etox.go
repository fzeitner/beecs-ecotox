package sys

import (
	"math/rand/v2"

	"github.com/mlange-42/ark-tools/resource"
	"github.com/mlange-42/ark/ecs"
	"github.com/mlange-42/beecs/comp"
	"github.com/mlange-42/beecs/globals"
	"github.com/mlange-42/beecs/params"
)

// InitEtox initializes and adds the resources
// necessary to simulate beecs_ecotox.
type InitEtox struct {
	larvaeEtox globals.LarvaeEtox
	inHiveEtox globals.InHiveEtox
	etox       *params.PPPApplication

	foragerFilter    *ecs.Filter1[comp.Activity]
	patchFilter      *ecs.Filter1[comp.Coords]
	source           rand.Source
	foragerPPPmapper *ecs.Map2[comp.PPPExpo, comp.EtoxLoad]
	patchPPPmapper   *ecs.Map2[comp.PatchPropertiesEtox, comp.ResourceEtox]
}

func (s *InitEtox) Initialize(w *ecs.World) {
	// initialize the globals for larvae/IHbee exposure
	aff := ecs.GetResource[params.AgeFirstForaging](w)
	workerDev := ecs.GetResource[params.WorkerDevelopment](w)
	droneDev := ecs.GetResource[params.DroneDevelopment](w)
	s.etox = ecs.GetResource[params.PPPApplication](w)

	if s.etox.FixedNectarPollenRatio {
		s.etox.PPPconcentrationPollen = s.etox.PPPconcentrationNectar * s.etox.EtoxNecPolFactor
	}

	s.larvaeEtox = globals.LarvaeEtox{
		WorkerCohortDose: make([]float64, workerDev.LarvaeTime),
		DroneCohortDose:  make([]float64, droneDev.LarvaeTime),
	}
	ecs.AddResource(w, &s.larvaeEtox)

	s.inHiveEtox = globals.InHiveEtox{
		WorkerCohortDose: make([]float64, aff.Max+1),
		DroneCohortDose:  make([]float64, droneDev.MaxLifespan),
	}
	ecs.AddResource(w, &s.inHiveEtox)

	// initialize ETOX storage globals
	init := ecs.GetResource[params.InitialStores](w)
	energyParams := ecs.GetResource[params.EnergyContent](w)
	storagesEtox := globals.StoragesEtox{
		EtoxHoneyEnergy:        make([]float64, 6),
		EtoxHoneyConcentration: make([]float64, 6),
	}
	storagesEtox.EtoxHoneyEnergy[5] = init.Honey * 1000.0 * energyParams.Honey
	ecs.AddResource(w, &storagesEtox)

	PPPfate := globals.PPPFate{}
	ecs.AddResource(w, &PPPfate)

	statsEtox := globals.PopulationStatsEtox{}
	ecs.AddResource(w, &statsEtox)

	// add the PPPExpo component to all foragers
	s.source = rand.New(ecs.GetResource[resource.Rand](w))
	s.foragerPPPmapper = s.foragerPPPmapper.New(w)
	s.foragerFilter = s.foragerFilter.New(w)

	query := s.foragerFilter.Query()
	toAdd := []ecs.Entity{}

	for query.Next() {
		toAdd = append(toAdd, query.Entity())
	}

	rng := rand.New(s.source)
	for _, entity := range toAdd {
		s.foragerPPPmapper.Add(entity, &comp.PPPExpo{OralDose: 0., ContactDose: 0., RdmSurvivalContact: rng.Float64(), RdmSurvivalOral: rng.Float64()}, &comp.EtoxLoad{PPPLoad: 0.})
	}
	toAdd = toAdd[:0]

	// add the PPP components to all patches
	s.patchPPPmapper = s.patchPPPmapper.New(w)
	s.patchFilter = s.patchFilter.New(w)

	pquery := s.patchFilter.Without(ecs.C[comp.ResourceEtox]()).Query()
	for pquery.Next() {
		toAdd = append(toAdd, pquery.Entity())
	}
	for _, entity := range toAdd {
		s.patchPPPmapper.Add(entity, &comp.PatchPropertiesEtox{PPPconcentrationNectar: 0., PPPconcentrationPollen: 0., PPPcontactDose: 0.}, &comp.ResourceEtox{PPPconcentrationNectar: 0., PPPconcentrationPollen: 0., PPPcontactDose: 0.})
	}
}

func (s *InitEtox) Update(w *ecs.World) {}

func (s *InitEtox) Finalize(w *ecs.World) {}
