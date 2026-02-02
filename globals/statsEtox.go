package globals

// PopulationStatsEtox contains the mean and cumulative oral doses of various cohorts.
//
// PopulationStatsEtox is updated at the end of each simulation step.
// Thus, it contains stats of the previous step.
type PopulationStatsEtox struct {
	NumberIHbeeCohorts int // only for debugging, will probably remove this later

	MeanDoseIHBees      float64 // changed to a mean dose per bee; doesn´t do anything, just there for debugging
	MeanDoseLarvae      float64 // changed to a mean dose per bee; doesn´t do anything, just there for debugging
	MeanDoseDrones      float64 // changed to a mean dose per bee; doesn´t do anything, just there for debugging
	MeanDoseDroneLarvae float64 // changed to a mean dose per bee; doesn´t do anything, just there for debugging
	MeanDoseForager     float64 // changed to a mean dose per bee; doesn´t do anything, just there for debugging

	CumDoseIHBees      float64 // cumulative dose before calculating a mean, used for debugging
	CumDoseLarvae      float64 // cumulative dose before calculating a mean, used for debugging
	CumDoseForagers    float64 // cumulative dose before calculating a mean, used for debugging
	CumDoseDrones      float64 // cumulative dose before calculating a mean, used for debugging
	CumDoseDroneLarvae float64 // cumulative dose before calculating a mean, used for debugging

	PPPNursebees float64 // variable for debugging and finding out how much PPP is "lost" to nursebees, who are not explicitely modeled

	CumDoseNurses  float64 // changed to a mean dose per bee; doesn´t do anything, just there for debugging
	MeanDoseNurses float64 // cumulative dose before calculating a mean, used for debugging
}

// Reset all stats to zero.
func (s *PopulationStatsEtox) Reset() {
	s.MeanDoseIHBees = 0      // original model actually only calculates the exposure per cohort and divides by number of individualy per cohort for mean doses
	s.MeanDoseLarvae = 0      // original model actually only calculates the exposure per cohort and divides by number of individualy per cohort for mean doses
	s.MeanDoseDrones = 0      // original model actually only calculates the exposure per cohort and divides by number of individualy per cohort for mean doses
	s.MeanDoseDroneLarvae = 0 // original model actually only calculates the exposure per cohort and divides by number of individualy per cohort for mean doses

	s.CumDoseIHBees = 0      // cumulative dose before calculating a mean, used for debugging
	s.CumDoseLarvae = 0      // cumulative dose before calculating a mean, used for debugging
	s.CumDoseDrones = 0      // cumulative dose before calculating a mean, used for debugging
	s.CumDoseDroneLarvae = 0 // cumulative dose before calculating a mean, used for debugging

	s.CumDoseNurses = 0
	s.MeanDoseNurses = 0

	s.PPPNursebees = 0

	s.CumDoseNurses = 0
	s.MeanDoseNurses = 0

}
