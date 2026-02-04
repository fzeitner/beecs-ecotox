package sys

import (
	"math"

	"github.com/mlange-42/ark-tools/resource"
	"github.com/mlange-42/ark/ecs"
	"github.com/mlange-42/beecs/comp"
	"github.com/mlange-42/beecs/globals"
	"github.com/mlange-42/beecs/params"
)

// MortalityForagersEtox applies worker mortality, including
//   - mortality from PPP exposure if applicable
//
// The proc is functionally identical to the NetLogo version.
type MortalityForagersEtox struct {
	rng                  *resource.Rand
	toRemove             []ecs.Entity
	foragerFilter        *ecs.Filter1[comp.PPPExpo]
	foragersFilterSimple *ecs.Filter0

	etoxStats *globals.PopulationStatsEtox
	etox      *params.PPPApplication
	toxic     *params.PPPToxicity
}

func (s *MortalityForagersEtox) Initialize(w *ecs.World) {
	s.rng = ecs.GetResource[resource.Rand](w)
	s.foragerFilter = s.foragerFilter.New(w)
	s.foragersFilterSimple = ecs.NewFilter0(w).With(ecs.C[comp.Age]())

	s.etoxStats = ecs.GetResource[globals.PopulationStatsEtox](w)
	s.etox = ecs.GetResource[params.PPPApplication](w)
	s.toxic = ecs.GetResource[params.PPPToxicity](w)
}

func (s *MortalityForagersEtox) Update(w *ecs.World) {
	query := s.foragerFilter.Query()
	s.etoxStats.MeanDoseForager = 0.
	s.etoxStats.CumDoseForagers = 0.
	c := query.Count()

	for query.Next() {
		p := query.Get()
		// mortality from PPP exposition, dose-response relationship depending on their susceptibility to the contaminant; theoretically BeeGUTS can also be implemented and called here
		lethaldose := false
		if s.etox.Application {
			s.etoxStats.CumDoseForagers += p.OralDose * 100
			if p.OralDose > 1e-20 && p.OralDose < s.toxic.ForagerOralLD50*1e5 {
				if p.RdmSurvivalOral < 1-(1/(1+math.Pow(p.OralDose/s.toxic.ForagerOralLD50, s.toxic.ForagerOralSlope))) {
					lethaldose = true
				}
			}
			if p.ContactDose > 0 {
				if p.RdmSurvivalContact < 1-(1/(1+math.Pow(p.ContactDose/s.toxic.ForagerContactLD50, s.toxic.ForagerContactSlope))) {
					lethaldose = true
				}
			}
			p.OralDose = 0.    // exposure doses get reset to 0 every tick BEFORE the added dose from honey and pollen consumption gets taken into account,
			p.ContactDose = 0. // therefore exposure from foraging of the current day and exposure from food of the previous day is relevant for lethal effects only
		}
		if lethaldose {
			s.toRemove = append(s.toRemove, query.Entity())
		}
	}

	for _, e := range s.toRemove {
		w.RemoveEntity(e)
	}
	s.toRemove = s.toRemove[:0]

	if c > 0 {
		s.etoxStats.MeanDoseForager = s.etoxStats.CumDoseForagers / float64(c*100)
	} else {
		s.etoxStats.MeanDoseForager = 0.
	}
}

func (s *MortalityForagersEtox) Finalize(w *ecs.World) {}
